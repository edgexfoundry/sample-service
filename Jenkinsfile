loadGlobalLibrary()

def BUILD_NODE = env.BUILD_NODE ?: 'centos7-docker-4c-2g'

node(BUILD_NODE) {
    stage('👭 Clone 👬') {
        edgeXScmCheckout()
        sh 'env | sort'
    }

    //////////////////////////////////////////////////////////////////////
    // {project-name}-verify-pipeline
    //////////////////////////////////////////////////////////////////////

    stage('🍳 Prep Builder') {
        def buildArgs = [
            '-f docker/Dockerfile',
            '.'
        ]
        buildImage = docker.build("go-builder:${GIT_BRANCH_CLEAN}", buildArgs.join(' '))
    }

    stage('💉 Test') {
        buildImage.inside('-u 0:0') {
            sh 'make test'
        }
    }

    //////////////////////////////////////////////////////////////////////
    // {project-name}-merge-pipeline
    //////////////////////////////////////////////////////////////////////

    edgeXMergeStage {
        stage('Semver Init') {
            edgeXSemver 'init'

            //set the version number on the environment
            def semverVersion = edgeXSemver()
            env.setProperty('VERSION', semverVersion)
        }

        // This will create a local tag with the current version
        stage('🏷️ Semver Tag') {
            edgeXSemver('tag')
        }

        stage('🖋️ Mock Sigul Signing') {
            sh 'echo lftools sigul branch v${VERSION}'
            sh 'echo lftools sigul docker v${VERSION}'
        }

        // Stage artifacts on Nexus ???
        stage('📦 Mock Upload Artifact') {
            sh 'echo docker tag edgexfoundry/device-sdk-go:${VERSION}'
            sh 'echo docker push edgexfoundry/device-sdk-go:${VERSION}'
        }

        stage('⬆️ Semver Bump Patch Version') {
            edgeXSemver('bump patch')
            edgeXSemver('-push')
        }
    }

    edgeXPRStage {
        stage('Non-Release Branch or PR') {
            if(env.ghprbActualCommit) {
                println "Triggered by GHPRB plugin doing extra stuff maybe?"

                if(env.ghprbCommentBody != "null") {
                    if(env.ghprbCommentBody =~ /^recheck$/) {
                        //No semver functions on recheck
                        echo 'Recheck'
                    }
                }
                else {
                    //No semver stuff on new pr or push?
                }
            }
        }
    }
}

def loadGlobalLibrary() {
    library(identifier: 'edgex-global-pipelines@master', 
        retriever: legacySCM([
            $class: 'GitSCM',
            userRemoteConfigs: [[url: 'https://github.com/edgexfoundry-holding/edgex-global-pipelines.git']],
            branches: [[name: '*/master']],
            doGenerateSubmoduleConfigurations: false,
            extensions: [[
                $class: 'SubmoduleOption',
                recursiveSubmodules: true,
            ]]]
        )
    ) _
}
