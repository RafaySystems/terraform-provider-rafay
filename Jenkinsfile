// pipeline {
//   agent { label 'node1' }
//   environment {
//     DOCKER_BUILDKIT='1'
//   }
//   stages {
//     stage('Build and push to S3') {
//       steps {
//         withCredentials([usernamePassword(credentialsId: 'jenkinsrafaygithub', passwordVariable: 'passWord', usernameVariable: 'userName')]) {
//           sh 'make release BUILD_USER=$userName BUILD_PASSWORD=$passWord'
//           sh 'make push BUILD_USER=$userName BUILD_PASSWORD=$passWord'
//         }
//       }
//     }
//   }
// }

pipeline {
    agent {
        docker { image 'public.ecr.aws/bitnami/golang:1.16.4' }
    }
    stages {
        stage('Build and push to S3') {
            steps {
                sh '''
                go version
                make release
                make push
                '''
            }
        }
    }
}