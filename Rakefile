require 'fileutils'
require './build'

ENV['GOPATH'] = [
	Dir.pwd,
	File.join(Dir.pwd, 'vendor'),
].join(':')

SRCS = FileList['src/sonar/**/*']
PROGS = ['vendor/bin/go-bindata', 'vendor/bin/pork']

DEPS = SRCS + PROGS + ['src/sonar/web/internal/bindata.go']

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

PROGS.each do |f|
  	file! f => ['bin/grr'] do
    	sh 'bin/grr', 'install'
  	end
end

file! 'src/sonar/web/internal/bindata.go' => PROGS + FileList['src/sonar/web/assets/**/*'] do |t|
  sh 'vendor/bin/pork',
  		'build',
  		'--out=dst/pub',
  		'--opt=basic',
  		'src/sonar/web/assets'

  sh 'rsync',
  		'-r',
  		'--exclude=*.ts',
  		'--exclude=*.scss',
  		'src/sonar/web/assets/',
  		'dst/pub'

  sh 'vendor/bin/go-bindata', '-o', t.name,
    	'-pkg', 'internal',
    	'-prefix', 'dst/pub',
    	'-ignore', '(\.ts|\.scss)$',
    	'dst/pub'
end

file! 'bin/grr' do
  	sh 'go', 'get', '-u', 'github.com/kellegous/grr'
  	FileUtils::rm_rf('src/github.com')
end

task :atom do
  	sh 'atom', '.'
end

task :subl do
  	sh 'subl', 'sonar.sublime-project'
end

task :default => CMDS

task :test do
	sh 'go', 'test',
		'sonar/config',
		'sonar/store'
end
