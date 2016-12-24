require 'rake'

TOCLEAN = FileList[]

def file!(*args, &block)
  task = Rake::FileTask.define_task(*args, &block)
  TOCLEAN.include(task.name)
  task
end

task :clean do
  TOCLEAN.each do |f|
    rm_rf(f) rescue nil
  end
end

def set_gopath(paths)
  ENV['GOPATH'] = paths.map { |p|
    "#{Dir.pwd}/#{p}"
  }.join(':')
end

def go_get(dst, deps)
	deps.map do |pkg|
		path = pkg.gsub(/\/\.\.\.$/, '')
		dest = File.join(dst, path)
		file dest do
			sh 'go', 'get', pkg
		end
		dest
	end
end

class Path
  def <<(p)
    path = ENV['PATH']
    p = File.join(Dir.pwd, p)
    ENV['PATH'] = "#{p}:#{path}"
  end
end

def path
  Path.new
end

# generates build rules for protobufs. Rules that target dst are generated
# by scanning src.
def protoc(src)
  FileList["#{src}/**/*.proto"].map do |src_path|
    dst_path = src_path.sub(/\.proto/, '.pb.go')
    file dst_path => [src_path] do
      sh 'protoc', "-I#{src}", src_path, "--go_out=#{src}"
    end

    dst_path
  end
end
