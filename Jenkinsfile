//
// Copyright (c) 2019 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
@Library("edgex-global-pipelines@experimental") _

def nginxPassthrough = '''worker_processes 2;
error_log nxginx_error.log;
events {
    worker_connections 1024;
}

http {
    server {
        listen 8080;
        server_name localhost;

        location /v2/ {
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_connect_timeout 150;
            proxy_hide_header WWW-Authenticate; #this header seems to break the implementation

            proxy_send_timeout 100;
            proxy_read_timeout 100;
            proxy_pass https://nexus3.edgexfoundry.org/repository/docker.io/v2/;
        }

    }
}
'''

pipeline {
    agent { label 'centos7-docker-4c-2g' }
    environment {
        DOCKER_PROXY_DIR = '/tmp/docker-proxy'
    }
    stages {
        stage('Docker') {
            steps {
                sh 'mkdir -p $DOCKER_PROXY_DIR'
                writeFile file: "/tmp/${env.DOCKER_PROXY_DIR}/nexus.conf" text: nginxPassthrough
                // spin up nginx container
                sh 'docker run --rm -p 8080:8080 -v $DOCKER_PROXY_DIR/nexus.conf:/etc/nginx/nginx.conf nginx:latest /bin/bash -c "cat /etc/nginx/nginx.conf && nginx -g \'daemon off;\'"'
                sh 'docker ps -a'

                // setup docker daemon

                sh 'sudo cat /etc/docker/daemon.json'
                sh 'jq '. + {"registry-mirrors": "http://localhost:8080", debug: true}' /etc/docker/daemon.json > /etc/docker/daemon.json'
                sh 'sudo cat /etc/docker/daemon.json'
                sh 'sudo service docker restart'

                sh 'docker pull alpine:3.10'
                sh 'sudo tail -200 /var/log/messages'
            }
        }
    }
}