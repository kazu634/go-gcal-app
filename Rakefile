desc 'Install the initial `go` binaries'
task :setup do
  sh 'go get github.com/Masterminds/glide'
  sh 'go get github.com/golang/lint/golint'
  sh 'go get golang.org/x/tools/cmd/goimports'
end

desc 'Install dependencies'
task deps: [ :setup ] do
  sh 'glide install'
end

desc 'Update dependencies'
task update: [ :setup ] do
  sh 'glide update'
end

desc 'Format source codes'
task fmt: [ :setup ] do
  sh 'goimports -w $(glide nv -x)'
end

desc 'Lint'
task lint: [ :setup ] do
  `glide novendor -x`.split.each do |target|
    sh "golint -set_exit_status #{target} || exit $?"
  end
end

desc 'Build binary'
task :build do
  sh 'git rev-parse --is-inside-work-tree' do |ok, status|
    if ok
      version = `git describe --tags --abbrev=0`.chomp
      revision = `git rev-parse --short HEAD`.chomp
    else
      version = '0.0'
      revision = 'xxxxxxxx'
    end

    ldflags = "-X main.version=#{version} -X main.revision=#{revision}"

    sh "go build -ldflags \"#{ldflags}\" main.go"
  end
end
