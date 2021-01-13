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

def swaggerScript = '''
#!/bin/bash
shopt -s extglob
# Environment Variables need to be set to call this shell script:
# Owner: The username on swaggerhub where this API will be pushed.
# apiFolder: A string of space delimited paths to API folders. Each .yaml file inside of this folder will be pushed to SwaggerHub.
# e.g.

publishToSwagger() {
    apiKey=$1
    apiFolder=$2
    oasVersion=$3
    isPrivate=$4
    owner=$5
    dryRun=${6:-false}

    apiPath="$WORKSPACE/${apiFolder}"

    echo "[publishToSwagger] Publishing the API Docs [${apiFolder}] to Swagger"

    if [ -d "$apiPath" ]; then
        for file in "${apiPath}"/*.+(yml|yaml); do
            apiName=$(basename "${file}" | cut -d "." -f 1)

            echo "[publishToSwagger] Publishing API Name [$apiName] [$file]"

            if [ "$dryRun" == "false" ]; then
                curl -v -X POST -d "@${file}" \
                    -H "accept:application/json" \
                    -H "Authorization:${apiKey}" \
                    -H "Content-Type:application/yaml" \
                    "https://api.swaggerhub.com/apis/${owner}/${apiName}?oas=${oasVersion}&isPrivate=${isPrivate}&force=true"
            else
                echo "[publishToSwagger] Dry Run enabled...Simulating upload"
                echo "curl -X POST https://api.swaggerhub.com/apis/${owner}/${apiName}?oas=${oasVersion}&isPrivate=${isPrivate}&force=true"
            fi
        done
    else
        echo "Could not find API Folder [${apiPath}]. Please make sure the API version exists..."
        exit 1
    fi
}

set -ex -o pipefail

echo "--> edgex-publish-swagger.sh"

# if no ARCH is set or ARCH is not arm
if [ -z "$ARCH" ] || [ "$ARCH" != "arm64" ] ; then

    # NOTE: APIKEY needs to be a pointer to a file with the key. This will need to be set locally from your environment or from Jenkins
    APIKEY_VALUE=$(cat "$APIKEY")

    # Upload all .yaml from within target to SwaggerHub
    SWAGGER_DRY_RUN=${SWAGGER_DRY_RUN:-false}
    OASVERSION='3.0.0'
    ISPRIVATE=false

    for API_FOLDER in ${API_FOLDERS}; do
        echo "=== Publish ${API_FOLDER} API ==="
        publishToSwagger "${APIKEY_VALUE}" "${API_FOLDER}" "${OASVERSION}" "${ISPRIVATE}" "${OWNER}" "${SWAGGER_DRY_RUN}"
    done

else
    echo "$ARCH not supported...skipping."
fi
'''

pipeline {
    agent { label 'centos7-docker-8c-8g' }
    stages {
        stage('Swagger Test') {
            steps {
                script {
                    git changelog: false, credentialsId: 'edgex-jenkins-access-username', poll: false, url: 'https://github.com/edgexfoundry/edgex-go.git'
                    writeFile file: 'edgex-publish-swagger.sh', text: swaggerScript
                    sh 'chmod +x edgex-publish-swagger.sh'

                    withEnv(["SWAGGER_DRY_RUN=false",
                            "OWNER=EdgeXFoundry1",
                            "API_FOLDERS=openapi/v1 openapi/v2"
                    ]) {
                        configFileProvider([configFile(fileId: 'swaggerhub-api-key', variable: 'APIKEY')]) {
                            sh './edgex-publish-swagger.sh'
                        }
                    }
                }
            }
        }
    }
}

