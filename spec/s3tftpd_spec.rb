require_relative 'spec_helper'

describe 's3tftpd' do
  LINE = /.*\n/

  def spawn_tftp
    PTY.spawn('tftp', 'localhost', port.to_s) do |cout, cin, pid|

      yield [cout, cin]
    end
  end

  let(:bucket_name) { ENV.fetch('TEST_BUCKET_NAME') }
  let(:command_args) { ["s3://#{bucket_name}"] }

  around do |example|
    activation = SocketActivation.new(Pathname(__dir__) + '../s3tftpd', *command_args)
    activation.add_fd(:udp)

    activation.run do |pid, (addr)|
      @port = addr.ip_port
      example.call
    end
  end

  let(:port) { @port }

  describe 'RRQ' do
    shared_examples 'success' do
      it 'sends correct data' do
        spawn_tftp do |cout, cin|
          tmp = Tempfile.new

          cout.expect!('tftp> ', 3)
          cin.puts("get #{request_path} #{tmp.path}")
          cout.expect!(LINE, 3) # discard echoback
          cout.expect!('tftp> ', 3)
          content = tmp.read
          expect(content).to match(expected_content)
        end
      end
    end

    shared_examples 'not-found' do
      it 'returns error' do
        spawn_tftp do |cout, cin|
          tmp = Tempfile.new

          cout.expect!('tftp> ', 3)
          cin.puts("get #{request_path} #{tmp.path}")
          cout.expect!(LINE, 3) # discard echoback
          expect(cout.expect!(LINE, 3)).to match([include('Error code 1')])
        end
      end
    end

    context 'get /test1' do
      include_examples 'success' do
        let(:request_path) { '/test1' }
        let(:expected_content) { "test object 1\n" }
      end
    end

    context 'get test1' do
      include_examples 'success' do
        let(:request_path) { 'test1' }
        let(:expected_content) { "test object 1\n" }
      end
    end

    context 'get /prefix/test2' do
      include_examples 'success' do
        let(:request_path) { '/prefix/test2' }
        let(:expected_content) { "test object 2\n" }
      end
    end

    context 'get prefix/test2' do
      include_examples 'success' do
        let(:request_path) { 'prefix/test2' }
        let(:expected_content) { "test object 2\n" }
      end
    end

    context 'When S3_URI has prefix' do
      let(:command_args) { ["s3://#{bucket_name}/prefix"] }

      context 'get /test1' do
        include_examples 'not-found' do
          let(:request_path) { '/test1' }
        end
      end

      context 'get /test2' do
        include_examples 'success' do
          let(:request_path) { '/test2' }
          let(:expected_content) { "test object 2\n" }
        end
      end
    end

    context 'get /not-found' do
      context 'get /not-found' do
        include_examples 'not-found' do
          let(:request_path) { '/not-found' }
        end
      end
    end

  end

  describe 'WRQ' do
    let(:content) { "test object upload #{Time.now}\n" }
    let(:request_path) { "writable/#{Time.now.to_f}" }

    it 'updates S3 object' do
      spawn_tftp do |cout, cin|
        tmp = Tempfile.new
        tmp.write(content)
        tmp.flush

        cout.expect!('tftp> ', 3)
        cin.puts("put #{tmp.path} #{request_path}")
        cout.expect!(LINE, 3) # discard echoback
        cout.expect!('tftp> ', 3)

        s3 = Aws::S3::Client.new
        remote_content = s3.get_object(bucket: bucket_name, key: request_path).body.read
        expect(remote_content).to eq content
      end
    end

    context 'When not writable' do
      let(:request_path) { "nonwritable/test" }

      it 'returns error' do
        spawn_tftp do |cout, cin|
          tmp = Tempfile.new
          tmp.write(content)
          tmp.flush

          cout.expect!('tftp> ', 3)
          cin.puts("put #{tmp.path} #{request_path}")
          cout.expect!(LINE, 3) # discard echoback
          expect(cout.expect!(LINE, 3)).to match([include('Error code 1')])
        end
      end
    end
  end
end
