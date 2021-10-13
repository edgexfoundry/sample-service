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
@Library("edgex-global-pipelines@c722b3ef8feead7275cc35cd43eeb32a32c03285") _

edgeXBuildGoApp (
    project: 'sample-service',
    goVersion: '1.16',
    buildExperimentalDockerImage: true
)

// def sampleMake = '''.PHONY: clean
// clean:
// 	echo "so fresh so clean"
// '''

// pipeline {
//     agent none
//     stages {
//         stage('Test') {
//             parallel {
//                 stage('Docker Test x86_64') {
//                     agent {
//                         label 'centos7-docker-4c-2g'
//                     }
//                     steps {
//                         sh 'docker version'
//                         writeFile(file: 'Makefile', text: sampleMake)
//                         script {
//                             docker.image('alpine:3.14').inside('-u 0:0') {
//                                 sh 'apk add --update make'
//                                 sh 'make clean'
//                             }
//                         }
//                     }
//                 }
//                 stage('Docker Test arm64') {
//                     agent {
//                         label 'ubuntu18.04-docker-arm64-4c-16g'
//                     }
//                     steps {
//                         // Issue with alpine:3.14 and older docker versions
//                         // See: https://wiki.alpinelinux.org/wiki/Release_Notes_for_Alpine_3.14.0#faccessat2
//                         sh 'sudo curl -o /etc/docker/seccomp.json "https://raw.githubusercontent.com/moby/moby/master/profiles/seccomp/default.json"'
//                         sh 'sudo cat /etc/docker/seccomp.json'
//                         sh 'sudo sed -i \'s/"defaultAction": "SCMP_ACT_ERRNO"/"defaultAction": "SCMP_ACT_TRACE"/g\' /etc/docker/seccomp.json'
//                         sh 'sudo jq \'. += {"seccomp-profile": "/etc/docker/seccomp.json"}\' /etc/docker/daemon.json | sudo tee /etc/docker/daemon.new'
//                         sh 'sudo mv /etc/docker/daemon.new /etc/docker/daemon.json'
//                         sh 'sudo service docker restart'
//                         sh 'docker version'
//                         writeFile(file: 'Makefile', text: sampleMake)
//                         script {
//                             docker.image('alpine:3.14').inside('-u 0:0') {
//                                 sh 'apk add --update make'
//                                 sh 'make clean'
//                             }
//                         }
//                     }
//                 }
//             }
//         }
//     }
// }
