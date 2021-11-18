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
@Library("edgex-global-pipelines@df90e98809a5d0daaba397a280e172605c56d1dc") _

pipeline {
    agent { label 'centos7-docker-4c-2g' }
    stages {
        stage('SSH Key Test') {
            environment {
                SEMVER_BRANCH = 'main'
            }
            parallel {
                stage('foo-bar-1') {
                    steps {
                        edgeXSemver('init')
                        sh 'cat VERSION'
                    }
                }
                stage('foo-bar-2') {
                    steps {
                        edgeXSemver('init')
                        sh 'cat VERSION'
                    }
                }
                stage('foo-bar-3') {
                    steps {
                        edgeXSemver('init')
                        sh 'cat VERSION'
                    }
                }
            }
        }
    }
}
