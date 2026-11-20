pipeline {
    agent {
        docker {
                image 'registry-proxy.dev.rafay-edge.net/golang:1.26.0'
                args '-u root:sudo'
                reuseNode false
                label 'ec2-fleet-tf'
            }
    }
    stages {
        stage('Build and push to S3') {
            steps {
                // GitHub App credential: passWord is a short-lived installation token minted per build.
                // Installation tokens must be used with the fixed username 'x-access-token' (written to .netrc below).
                withCredentials([usernamePassword(credentialsId: 'ops-jenkins-gh-app', passwordVariable: 'passWord', usernameVariable: 'githubAppId')]) {
                withCredentials([[$class: 'AmazonWebServicesCredentialsBinding', accessKeyVariable: 'AWS_ACCESS_KEY_ID', credentialsId: 'jenkinsAwsUser', secretKeyVariable: 'AWS_SECRET_ACCESS_KEY']]) {
                sh '''
                    apt-get update && apt-get install -y unzip
                    curl "https://awscli.amazonaws.com/awscli-exe-linux-$(uname -m).zip" -o "awscliv2.zip"
                    unzip -q -o awscliv2.zip
                    bash ./aws/install
                    go version
                    aws --version
                    export GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore
                    # Write .netrc with xtrace off so the token never appears in the console
                    # (Jenkins runs sh with -x, which would print the expanded command).
                    # AWS_* are already exported by withCredentials, so no need to re-export them.
                    { set +x; } 2>/dev/null
                    trap 'rm -f ~/.netrc' EXIT
                    ( umask 077; printf 'machine github.com\\nlogin x-access-token\\npassword %s\\n' "$passWord" > ~/.netrc )
                    set -x
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
