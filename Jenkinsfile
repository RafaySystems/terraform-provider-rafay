pipeline {
    agent {
        docker { 
                image 'public.ecr.aws/bitnami/golang:1.23.4'
                args '-u root:sudo'
                reuseNode false
            }
    }
    stages {
        stage('Build and push to S3') {
            steps {
                withCredentials([usernamePassword(credentialsId: 'jenkinsrafaygithub', passwordVariable: 'passWord', usernameVariable: 'userName')]) {
                withCredentials([[$class: 'AmazonWebServicesCredentialsBinding', accessKeyVariable: 'AWS_ACCESS_KEY_ID', credentialsId: 'jenkinsAwsUser', secretKeyVariable: 'AWS_SECRET_ACCESS_KEY']]) {
                sh '''
                    curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
                    unzip -q -o awscliv2.zip
                    bash ./aws/install
                    go version
                    aws --version
                    export AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID}
                    export AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY}
                    export GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore
                    echo machine github.com login ${userName} password ${passWord} > ~/.netrc
                    chmod 400 ~/.netrc
                    GOPROXY="https://proxy.golang.org,direct" GOPRIVATE="github.com/RafaySystems/*" go mod download
                    make release
                    make push
                    make bucket-name
                '''
                }}
            }
        }
    }
    post {
        success {
            slackSend channel: "#build",
            color: 'good',
            message: "Build ${currentBuild.fullDisplayName} completed successfully."
        }
        failure {
            slackSend channel: "#build",
            color: 'RED',
            message: "Attention ${env.JOB_NAME} ${env.BUILD_NUMBER} has failed."
        }
        always {
                sh '''
                chown -R 1000:1000 .
                '''
                deleteDir()
                dir("${workspace}@tmp") {
                    deleteDir()
                }
                dir("${workspace}@script") {
                    deleteDir()
            }
        }
    }
}
