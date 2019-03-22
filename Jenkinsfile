node('centos7-docker-4c-2g') {
    stage('Clone') {
        def gitVars = checkout scm
        setupEnvironment(gitVars)
    }

    stage('Semver Test') {
        semver 'init'
    }

    // stage('Docker Login') {
    //     configFileProvider(
    //         [configFile(fileId: 'sandbox-settings', variable: 'MAVEN_SETTINGS')]) {
    //           // nexus docker login stuff here?
    //     }
    // }
}

def setupEnvironment(vars) {
    if(vars != null) {
        vars.each { k, v ->
            env.setProperty(k, v)
            if(k == 'GIT_BRANCH') {
                env.setProperty('SEMVER_BRANCH', v.replaceAll( /^origin\//, '' ))
                env.setProperty('GIT_BRANCH_CLEAN', v.replaceAll('/', '_'))
            }
        }
    }
}

def semver(command = null, credentials = 'edgex-jenkins-ssh', debug = true) {
    def semverCommand = [
       'git',
       'semver'
    ]

    if(debug) { semverCommand << '-debug' }
    if(command) { semverCommand << command }

    docker.image('ernestoojeda/git-semver:alpine').inside('-v /etc/ssh:/etc/ssh') {
        withEnv(['SSH_KNOWN_HOSTS=/etc/ssh/ssh_known_hosts']) {
            sshagent (credentials: [credentials]) {
                sh semverCommand.join(' ')
            }
        }
    }
}