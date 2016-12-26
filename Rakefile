require 'fileutils'
require './build'

ENV['GOPATH'] = Dir.pwd

ENV['PATH'] = "#{Dir.pwd}/bin:#{ENV['PATH']}"

SRCS = FileList['src/sonar/**/*']
DEPS = SRCS
CMDS = FileList['src/sonar/cmds/*'].map do |f|
  name = File.basename(f)
  dest = File.join('bin', name)
  file! dest => DEPS do |t|
  	cmd = ['go', 'install']
  	if ENV['dev'] == 'true'
  		dir = File.absolute_path('src/sonar/web/assets')
  		cmd << '-ldflags'
  		cmd << "-X main.assetsDir=#{dir}"
  	end
  	cmd << "sonar/cmds/#{name}"
    sh *cmd
  end
  dest
end

task :atom do
  sh 'atom', '.'
end

task :subl do
  sh 'subl', 'sonar.sublime-project'
end

task :test do
end

task :default => CMDS

task :test do
	sh 'go', 'test',
		'sonar/config',
		'sonar/store'
end
