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
// @Library("edgex-global-pipelines@experimental") _

pipeline {
    agent {
        node {
            label edgex.mainNode([:])
        }
    }
    options {
        timestamps()
        preserveStashes()
        quietPeriod(5) // wait a few seconds before starting to aggregate builds...??
        durabilityHint 'PERFORMANCE_OPTIMIZED'
        timeout(360)
    }
    triggers {
        issueCommentTrigger('.*^recheck$.*')
    }
    parameters {
        string(
            name: 'JobName',
            defaultValue: '',
            description: 'The job name')
        string(
            name: 'CommitId',
            defaultValue: '',
            description: 'The commitId in the code repository from where to initiate the build - should be used only if building via edgeXRelease')
    }
    stages {
        stage('Check Commit ID') {
            steps {
                script {
                    if(params.JobName) {
                        sh "echo Hey I got this Job Name [${params.JobName}]"
                    }

                    if(params.CommitId) {
                        sh "echo Hey I got this CommitId [${params.CommitId}]"
                        sh "echo git checkout ${params.CommitId}"
                    }
                }
            }
        }
    }
}