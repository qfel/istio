FROM alpine:3.10.0 AS build
RUN apk add --update --no-cache \
    curl \
    git

# curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt
ARG KUBECTL_VERSION=1.15.0
ENV KUBECTL_URL="https://storage.googleapis.com/kubernetes-release/release/v${KUBECTL_VERSION}/bin/linux/amd64/kubectl"
RUN curl --remote-name --silent --location "${KUBECTL_URL}" && \
    chmod +x kubectl && \
    mv kubectl /usr/local/bin/

ARG HELM_VERSION=2.14.1
ENV HELM_URL="https://get.helm.sh/helm-v${HELM_VERSION}-linux-amd64.tar.gz"
RUN curl --silent --location "${HELM_URL}" | tar xvz && \
    chmod +x linux-amd64/helm && \
    mv linux-amd64/helm /usr/local/bin/

ARG INSTALLER_VERSION=master
RUN git clone https://github.com/istio/installer.git
RUN git --work-tree=installer --git-dir=installer/.git checkout "${INSTALLER_VERSION}"

FROM alpine:3.10.0
RUN apk add --update --no-cache \
    bash
COPY --from=build /usr/local/bin/kubectl /usr/local/bin/helm /usr/local/bin/
COPY --from=build /installer /root/installer
COPY install.sh /root/install.sh
WORKDIR /root
ENTRYPOINT ["/root/install.sh"]
