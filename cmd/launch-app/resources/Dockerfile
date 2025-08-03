FROM debian:stable-slim 

RUN apt update && apt install -y apt-transport-https ca-certificates gnupg curl git lsb-release
                
RUN curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg && \
    echo "deb [arch=amd64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu focal stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null && \
    apt update && apt install -y docker-ce docker-ce-cli containerd.io && \
    usermod -aG docker root