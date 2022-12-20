pipeline {
    agent {
        docker { 
                image 'public.ecr.aws/bitnami/golang:1.18.4'
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
                    echo machine github.com login ${userName} password ${passWord} > ~/.netrc
                    chmod 400 ~/.netrc
                    GOPRIVATE="github.com/RafaySystems/*" go mod download
                    make release
                    make push
                    make bucket-name
                '''
                }}
            }
        }
    }
}