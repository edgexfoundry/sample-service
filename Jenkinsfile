node('centos7-docker-4c-2g') {
  stage('SSH-Agent Test') {
      sshagent (credentials: ['edgex-jenkins-ssh']) {
          sh "ssh -v git@github.com"
      }
  }
}