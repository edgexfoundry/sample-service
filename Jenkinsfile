node('centos7-docker-4c-2g') {
    stage('SSH-Agent Test') {
        semver()
    }
}

def semver(command, credentials = 'edgex-jenkins-ssh', debug = true) {
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