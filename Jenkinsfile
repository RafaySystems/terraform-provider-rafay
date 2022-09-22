pipeline {
  agent { label 'node1' }
  environment {
    DOCKER_BUILDKIT='1'
  }
  stages {
    stage('Build and push to S3') {
      steps {
        sh 'make release'
        sh 'make push'
      }
    }
  }
}