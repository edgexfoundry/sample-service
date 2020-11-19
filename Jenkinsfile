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

pipeline {
    agent { label 'ubuntu18.04-docker-arm64-4c-16g' }
    stages {
        stage('Docker') {
            steps {
                enableDockerProxy('https://nexus3.edgexfoundry.org:10001')

                sh 'docker pull alpine:3.10'
                sh 'sudo tail -200 /var/log/messages'
            }
        }
    }
}

def enableDockerProxy(proxyHost, debug = false) {
    sh 'sudo cat /etc/docker/daemon.json'
    sh "sudo jq \'. + {\"registry-mirrors\": [\"${proxyHost}\"], debug: ${debug}}\' /etc/docker/daemon.json > /tmp/daemon.json"
    sh 'sudo mv /tmp/daemon.json /etc/docker/daemon.json'
    sh 'sudo cat /etc/docker/daemon.json'
    sh 'sudo service docker restart | true'
    sh 'systemctl status docker.service | true'
    //sh 'sudo tail -200 /var/log/messages'
    sh 'sudo journalctl -xe'
}