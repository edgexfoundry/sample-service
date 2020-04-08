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

// edgeXBuildGoApp (
//     project: 'sample-service',
//     goVersion: '1.12'
// )

pipeline {
    agent {
        label 'centos7-docker-4c-2g'
    }
   environment {
       DRY_RUN = 'false'
   }
    options {
        timestamps()
        quietPeriod(5)
        durabilityHint 'PERFORMANCE_OPTIMIZED'
        timeout(360)
    }
    stages {
        stage('Execute') {
            steps {
                edgeXReleaseGitTag([
                    'name': 'sample-service',
                    'version': '1.2.3',
                    'releaseStream': 'master',
                    'repo': 'https://github.com/edgexfoundry/sample-service.git'
                ])
            }
        }
    }
}