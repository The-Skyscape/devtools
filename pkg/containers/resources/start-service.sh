echo "Starting container {{.Name}} with image {{.Image}}..." && \
echo "Checking if image exists locally..." && \
if ! docker images --format "{{`{{.Repository}}:{{.Tag}}`}}" | grep -q "^{{.Image}}$"; then \
  echo "Image {{.Image}} not found locally, pulling..." && \
  docker pull {{.Image}} || { echo "Failed to pull image {{.Image}}"; exit 1; }; \
else \
  echo "Image {{.Image}} already exists locally"; \
fi && \
mkdir -p {{dataDir}}/services/{{.Name}} && \
chmod -R 777 {{dataDir}}/services/{{.Name}} && \
docker rm -f {{.Name}} 2>/dev/null || true && \
echo "Creating container {{.Name}}..." && \
docker create \
  --name {{.Name}} \
  {{if .RestartPolicy}}--restart {{.RestartPolicy}}{{end}} \
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