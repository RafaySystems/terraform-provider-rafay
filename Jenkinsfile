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
        docker { 
                image 'public.ecr.aws/bitnami/golang:1.18.4'
                reuseNode false
            }
    }
    stages {
        stage('Build and push to S3') {
            steps {
                withCredentials([usernamePassword(credentialsId: 'jenkinsrafaygithub', passwordVariable: 'passWord', usernameVariable: 'userName')]) {
                sh '''
                    go version
                    curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
                    unzip awscliv2.zip
                    bash ./aws/install --update
                    make release
                    make push
                '''
                }
            }
        }
    }
}