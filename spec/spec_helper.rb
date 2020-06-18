require 'aws-sdk-s3'
require 'expect'
require 'pathname'
require 'pty'
require 'socket'
require 'tempfile'

class SocketActivation
  def initialize(*cmd)
    @cmd = cmd.map(&:to_s)
    @fds = []
  end

  def add_fd(proto, host = 'localhost')
    @fds << [proto, host]
  end

  def run(killsig: :SIGKILL, &block)
    redir = @fds.map.with_index(3) do |(proto, host), fd|
      case proto
      when :udp
        [fd, Addrinfo.udp(host, 0).bind]
      else
        fail "Unknown proto #{proto.inspect}"
      end
    end

    pid = fork do
      ENV['LISTEN_FDS'] = @fds.size.to_s
      ENV['LISTEN_PID'] = Process.pid.to_s

      exec(*@cmd, redir.to_h.merge(close_others: true))
    end

    addresses = redir.map do |_, io|
      addr = io.local_address
      io.close
      addr
    end

    block.call(pid, addresses)

  ensure
    if pid
      Process.kill(killsig, pid)
      Process.waitpid(pid)
    end
  end

end

class IO
  def expect!(pattern, timeout = 3)
    unless ret = expect(pattern, timeout)
      fail "Expected #{pattern.inspect}"
    end
    ret
  end
end
