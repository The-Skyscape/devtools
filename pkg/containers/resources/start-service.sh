mkdir -p {{dataDir}}/services/{{.Name}}
chmod -R 777 {{dataDir}}/services/{{.Name}}

docker create \
  --name {{.Name}} \
  {{if .Network}}--network {{.Network}}{{end}} \
  {{if .Privileged}}--privileged{{end}} \
  {{if .Entrypoint}}--entrypoint {{.Entrypoint}}{{end}} \
  {{range $key, $val := .Ports}}-p {{$key}}:{{$val}} {{end}} \
  {{range $key, $val := .Mounts}}-v {{$key}}:{{$val}} {{end}} \
  {{range $key, $val := .Env}}-e {{$key}}={{$val}} {{end}} \
  {{if .Privileged}}-v /var/run/docker.sock:/var/run/docker.sock{{end}} \
  {{.Image}} {{.Command}};

{{range $src, $dst := .Copied}}
docker cp {{$src}} {{$.Name}}:{{$dst}}
{{end}}

docker start {{.Name}}

# {{if .Privileged}}
# docker exec -i {{.Name}} bash <<'EOF'
# apt update 
# apt install -y apt-transport-https ca-certificates curl gnupg git lsb-release

# curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg && \
# echo "deb [arch=amd64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu focal stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null && \
# apt update && \
# apt install -y docker-ce docker-ce-cli containerd.io && \
# usermod -aG docker root
# EOF
# {{end}}