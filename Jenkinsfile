def buildNode = env.BUILD_NODE ?: 'centos7-docker-4c-2g'

library(identifier: 'edgex-global-pipelines@master', 
    retriever: legacySCM([
        $class: 'GitSCM',
        branches: [[name: '*/master']],
        doGenerateSubmoduleConfigurations: false,
        extensions: [[
            $class: 'SubmoduleOption',
            recursiveSubmodules: true,
        ]],
        userRemoteConfigs: [[url: 'git@github.com:ernestojeda/edgex-global-pipelines.git']]])
) _

library identifier: 'edgex-global-pipelines@master', retriever: legacySCM([$class: 'GitSCM', branches: [[name: '*/master']], doGenerateSubmoduleConfigurations: false, extensions: [[$class: 'SubmoduleOption', disableSubmodules: false, parentCredentials: false, recursiveSubmodules: true, reference: '', trackingSubmodules: false]], submoduleCfg: [], userRemoteConfigs: [[url: 'git@github.com:ernestojeda/edgex-global-pipelines.git']]])


node(buildNode) {
    sh 'uname -m'

    stage('👭 Clone 👬') {
        edgeXScmCheckout()
        sh 'env | sort'
    }
/*
    if(releaseStream(env.GIT_BRANCH)) {
        stage('Semver Init') {
            semver 'init'

            docker.image("ernestoojeda/git-semver:${env.ARCH}").inside {
                env.setProperty('VERSION', sh(script: 'git semver', returnStdout: true).trim())
            }
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
    if(releaseStream(env.GIT_BRANCH)) {
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
    */
}

// def setupEnvironment(vars) {
//     if(vars != null) {
//         vars.each { k, v ->
//             env.setProperty(k, v)
//             if(k == 'GIT_BRANCH') {
//                 env.setProperty('SEMVER_BRANCH', v.replaceAll( /^origin\//, '' ))
//                 env.setProperty('GIT_BRANCH_CLEAN', v.replaceAll('/', '_'))
//             }
//         }
//     }

//     // set default architecture
//     if(!env.ARCH) {
//         def vmArch = sh(script: 'uname -m', returnStdout: true).trim()
//         env.setProperty('ARCH', vmArch)
//     }

//     if(releaseStream(env.GIT_BRANCH)) {
//         semver 'init'

//         docker.image("ernestoojeda/git-semver:${env.ARCH}").inside {
//             env.setProperty('VERSION', sh(script: 'git semver', returnStdout: true).trim())
//         }
//     }
// }

def semver(command = null, credentials = 'edgex-jenkins-ssh', debug = true) {
    def semverCommand = [
       'git',
       'semver'
    ]

    if(debug) { semverCommand << '-debug' }
    if(command) { semverCommand << command }

    docker.image("ernestoojeda/git-semver:${env.ARCH}").inside('-v /etc/ssh:/etc/ssh') {
        withEnv(['SSH_KNOWN_HOSTS=/etc/ssh/ssh_known_hosts']) {
            sshagent (credentials: [credentials]) {
                sh semverCommand.join(' ')
            }
        }
    }
}

def releaseStream(branchName) {
    (getStreams().collect { branchName =~ it ? true : false }).contains(true)
}

def getStreams() {
    [/.*master/, /.*delhi/, /.*edinburgh/, /.*git-semver/]
}