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
    agent { label 'centos7-docker-4c-2g' }
    stages {
        stage('SSH Key Test') {
            environment {
                SSH_KNOWN_HOSTS='/etc/ssh/ssh_known_hosts'
                SEMVER_BRANCH = 'main'
            }
            steps {
                script {
                    setupKnownHosts()
                    docker.image('nexus3.edgexfoundry.org:10004/edgex-devops/git-semver:latest').inside('-u 0:0 -v /etc/ssh:/etc/ssh') {
                        sshagent (credentials: ['edgex-jenkins-ssh']) {
                            sh 'git semver init'
                            sh 'git semver'
                        }
                    }
                }
            }
        }
    }
}

def setupKnownHosts() {
    sh '''
    if ! grep "github.com ecdsa" /etc/ssh/ssh_known_hosts; then
        grep -v github /etc/ssh/ssh_known_hosts > /tmp/ssh_known_hosts
        if [ -e /tmp/ssh_known_hosts ]; then
            sudo mv /tmp/ssh_known_hosts /etc/ssh/ssh_known_hosts
            echo "github.com ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBEmKSENjQEezOmxkZMy7opKgwFB9nkt5YRrYMjNuG5N87uRgg6CLrbo5wAdT/y6v0mKV0U2w0WZ2YB/++Tpockg=" | sudo tee -a /etc/ssh/ssh_known_hosts
        fi
    fi
    '''
}