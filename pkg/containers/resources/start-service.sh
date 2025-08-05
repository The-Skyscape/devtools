mkdir -p {{dataDir}}/services/{{.Name}} && \
chmod -R 777 {{dataDir}}/services/{{.Name}} && \
docker rm -f {{.Name}} 2>/dev/null || true && \
docker create \
  --name {{.Name}} \
  {{if .Network}}--network {{.Network}}{{end}} \
  {{if .Privileged}}--privileged{{end}} \
  {{if .Entrypoint}}--entrypoint {{.Entrypoint}}{{end}} \
  {{range $key, $val := .Ports}}-p {{$key}}:{{$val}} {{end}} \
  {{range $key, $val := .Mounts}}-v {{$key}}:{{$val}} {{end}} \
  {{range $key, $val := .Env}}-e {{$key}}={{$val}} {{end}} \
  {{if .Privileged}}-v /var/run/docker.sock:/var/run/docker.sock{{end}} \
  {{.Image}} {{.Command}}{{range $src, $dst := .Copied}} && \
docker cp {{$src}} {{$.Name}}:{{$dst}}{{end}} && \
docker start {{.Name}} && \
echo "Container started successfully"