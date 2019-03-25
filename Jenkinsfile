loadGlobalLibrary()

def BUILD_NODE = env.BUILD_NODE ?: 'centos7-docker-4c-2g'
def RELEASE_STREAMS = [/.*master/, /.*delhi/, /.*edinburgh/, /.*git-semver/]

node(BUILD_NODE) {
    stage('👭 Clone 👬') {
        edgeXScmCheckout()
        sh 'env | sort'
    }

    if(isReleaseStream()) {
        stage('Semver Init') {
            edgeXSemver 'init'

            def semverVersion = edgeXSemver()
            env.setProperty('VERSION', semverVersion)
        }
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

    // Master branch
    if(isReleaseStream()) {
        // This will create a local tag with the current version
        stage('🏷️ Semver Tag') {
            semver('tag')
        }

        // Stage artifacts on Nexus ???
        stage('📦 Upload Artifact Mockup') {
            sh 'echo docker tag edgexfoundry/device-sdk-go:${VERSION}'
            sh 'echo docker push edgexfoundry/device-sdk-go:${VERSION}'
        }

        stage('🖋️ Mock Sigul Signing') {
            sh 'echo lftools sigul branch v${VERSION}'
            sh 'echo lftools sigul docker v${VERSION}'
        }

        stage('⬆️ Semver Bump Patch Version') {
            semver('bump patch')
            semver('-push')
        }
    }
    // everything else
    else {
        stage('Non-Release Branch or PR') {
            //if Using the GHPRB plugin
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
            branches: [[name: '*/master']],
            doGenerateSubmoduleConfigurations: false,
            extensions: [[
                $class: 'SubmoduleOption',
                recursiveSubmodules: true,
            ]],
            userRemoteConfigs: [[url: 'https://github.com/ernestojeda/edgex-global-pipelines.git']]])
    ) _
}

def isReleaseStream(branchName = env.GIT_BRANCH) {
    branchName
        ? (RELEASE_STREAMS.collect { branchName =~ it ? true : false }).contains(true)
        : false
}
