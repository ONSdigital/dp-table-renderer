#!groovy

node {
    def gopath = pwd()

    ws("${gopath}/src/github.com/ONSdigital/dp-table-renderer") {
        stage('Checkout') {
            checkout scm
            sh 'git clean -dfx'
            sh 'git rev-parse --short HEAD > git-commit'
            sh 'set +e && (git describe --exact-match HEAD || true) > git-tag'
        }

        def branch   = env.JOB_NAME.replaceFirst('.+/', '')
        def revision = revisionFrom(readFile('git-tag').trim(), readFile('git-commit').trim())

        stage('Build') {
            sh "GOPATH=${gopath} make"
        }

        stage('Test') {
            sh "GOPATH=${gopath} make test"
        }
    }
}

@NonCPS
def revisionFrom(tag, commit) {
    def matcher = (tag =~ /^release\/(\d+\.\d+\.\d+(?:-rc\d+)?)$/)
    matcher.matches() ? matcher[0][1] : commit
}
