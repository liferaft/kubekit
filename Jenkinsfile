@Library('kubekit-shared-ci-library@master') _
def server = Artifactory.server 'jenkins_artifactory_server'

DEFAULT_PIPELINE_NODE     = 'openstack' // vsphere/ubuntu(openstack)
DEFAULT_BUILD_NODE     = 'openstack' // vsphere/ubuntu(openstack)
//This file can't be larger than 32000 bytes or you will hit Method code too large!
pipeline {
  // Don't use a heavyweight executor at this level to avoid blocking an
  // agent since we'll be asking for user input at some stages.
  agent { node { label DEFAULT_PIPELINE_NODE as String } }

  parameters {
    booleanParam(name: 'CRON_TRIGGER', defaultValue: false, description: 'Was this job kicked off via cron')
    booleanParam(name: 'EC2_ENABLED', defaultValue: false, description: 'Are we going to run on EC2:')
    booleanParam(name: 'EKS_ENABLED', defaultValue: false, description: 'Are we going to run on EKS:')
    booleanParam(name: 'OPENSTACK_ENABLED', defaultValue: false, description: 'Are we going to run on Openstack:')
    booleanParam(name: 'VSPHERE_ENABLED', defaultValue: true, description: 'Are we going to run on vSphere:')
    booleanParam(name: 'OPENSTACK_HARDENED_ENABLED', defaultValue: false, description: 'Are we going to run on Openstack HARDENED:')
    booleanParam(name: 'OPENSTACK_SLES15_ENABLED', defaultValue: false, description: 'Are we going to run on Openstack SLES15:')
    booleanParam(name: 'AKS_ENABLED', defaultValue: false, description: 'Are we going to run on AKS:')
    booleanParam(name: 'BLACKDUCK_ENABLED', defaultValue: false, description: 'Force Blackduck to run')
    booleanParam(name: 'ENABLE_DEBUG', defaultValue: false, description: 'Enable debug output in kubekit')
    booleanParam(name: 'FORCE_RPM_BUILD', defaultValue: false, description: 'Are we going to force an PR rpm build')

    string(name: 'VENDOR_UPDATE', defaultValue: '', description: 'Which Project triggering the build')
    string(name: 'TAR_COMMENT', defaultValue: '', description: 'String to add to tar file name')
  }

  options {
      timestamps()
      buildDiscarder(logRotator(daysToKeepStr:'120'))
      ansiColor('xterm')
      timeout(time: 2, unit: 'HOURS')
  }

  environment {
    OUTPUT_DIR    = 'pkg'
    SNAPSHOT_REPO = 'dependencies-snapshot-sd'
    QA_REPO       = 'dependencies-qa-sd'
    STABLE_REPO   = 'dependencies-stable-sd'
    RELEASED_REPO = 'dependencies-released-sd'
    REPO_PATH     = 'shared-services/liferaft/kubekit'
    golang_path   = "/go/src/github.com/liferaft/kubekit"
    OPENSTACK_HARDENED = "bb86a640-1b31-40d5-a042-60485ee4020a"
    OPENSTACK_SLES15 = "d421f2c4-1569-490e-aad6-cfc10f2b2e30"
  }

  triggers { parameterizedCron('''H 15 * * * % CRON_TRIGGER=true''') }

  stages {
    stage ('Prep') {
      agent { node { label DEFAULT_PIPELINE_NODE as String } }
      steps {
        script {
          if ( env.CHANGE_TARGET ) { // This is a PR
            env.BRANCH = CHANGE_TARGET
          } else {
            env.BRANCH=BRANCH_NAME
            if ( params.CRON_TRIGGER ) { echo "PICKLE RICK!!!!!! $CRON_TRIGGER" }//if
          }//else
          env.VERSION = sh(returnStdout: true, script: "make version SHORT=true").trim()
          env.DEFAULT_KUBEKIT_BIN='kubekit_linux_amd64'
          env.gitCommit = sh(returnStdout:true, script:"git rev-parse --short=7 HEAD").trim()
          env.DEV_VERSION = "${env.VERSION}.${BUILD_TIMESTAMP}.${gitCommit}.${env.BUILD_ID}".trim()
          env.GIT_LOG_MESSAGE = sh(returnStdout:true, script:"git shortlog origin/${BRANCH}..$gitCommit").toLowerCase().trim().replaceAll("[\n\r]", "")
          env.GIT_COMMIT_FILES = "" // sh(returnStdout:true, script:"git diff origin/${BRANCH}..$gitCommit --name-only").toLowerCase().trim().replaceAll("[\n\r]", "")
          env.GOLANG = getGolangVer()
          env.BASE_MACHINE_NAME = "${env.gitCommit}-${env.BUILD_ID}"
          if ( env.CHANGE_BRANCH ) {
            env.KK_PATH = "${SNAPSHOT_REPO}/*/${BRANCH_NAME}-${DEV_VERSION}"
          } else {
            env.KK_PATH = "${SNAPSHOT_REPO}/*/${DEV_VERSION}"
          }
        }//script
        decorateBuild( version: "${env.VERSION}", git_log_message: "${GIT_LOG_MESSAGE}", golang: "{$GOLANG}" )
        sh(script: 'env | sort')
        // DEBUG: remove this to speed up pipeline
      }//steps
    } //stage prep

    stage('build dependencies') {
      agent { node { label DEFAULT_BUILD_NODE as String } }
      steps{
        BuildDockerImage()
      }
      post { always { postAction(cleanws: true) } }//post
    }
    stage('Parallel tests') {
      parallel {
        stage ('BlackDuck Test') {
          agent { node { label DEFAULT_BUILD_NODE as String } }
          steps {
            BlackDuck(projectVersion:"${env.VERSION}-${env.BRANCH_NAME}",
                      url:"https://bdhub.com",
                      forceSuccess:"true"
                      )
          } //steps
          post { always { postAction(cleanws: true) } } //post
        } //stage blackduck
        // DEBUG: remove this to speed up pipeline
        stage('Go fmt') {
          agent { node { label DEFAULT_BUILD_NODE as String } }
          steps {
            BuildGoInDockerImage( makeopts: "fmt" )
          } //steps
          post { always { postAction(cleanws: true) } }//post
        } //stage
        // DEBUG: remove this to speed up pipeline
        stage('Code coverage tests') {
          agent { node { label DEFAULT_BUILD_NODE as String } }
          steps {
            sh(script: """mkdir -p ${WORKSPACE}/go-build """)
            withCredentials([
              usernamePassword(
                credentialsId: "aws_service_account",
                usernameVariable: 'PLATFORM_USERID',
                passwordVariable: 'PLATFORM_PASSWORD')
              ]) {
                BuildGoInDockerImage(
                  makeopts: " coverage",
                  dockerOpts: "-e AWS_ACCESS_KEY_ID=${PLATFORM_USERID} -e AWS_SECRET_ACCESS_KEY=${PLATFORM_PASSWORD} -v ${WORKSPACE}/go-build:/go/bin "
                )
            } //withCredentials
            sh(script: " sed -i \"s~<source>.*</source>~<source>${WORKSPACE}</source>~g\" ${WORKSPACE}/cobertura.xml " )
            cobertura autoUpdateHealth: false, autoUpdateStability: false, coberturaReportFile: 'cobertura.xml', conditionalCoverageTargets: '70, 0, 0', failNoReports: false, failUnhealthy: false, failUnstable: false, lineCoverageTargets: '80, 0, 0', maxNumberOfBuilds: 0, methodCoverageTargets: '80, 0, 0', sourceEncoding: 'ASCII', zoomCoverageChart: false
          } //steps
          post { always { postAction(cleanws: true) } }//post
        } //stage
        stage ('Package Build') {
          agent { node { label DEFAULT_BUILD_NODE as String } }
          steps {
            script {
              if ( findPlatform( keywords: "(?i)(.*)pkg/manifest/release(.*)") || FORCE_RPM_BUILD ) {
                if ( env.CHANGE_BRANCH ) { //This is a PR, build new RPM using this branch
                  haveChangedManifest = sh(returnStdout:true, script:"git diff --name-only HEAD~1").trim().replaceAll("[\n\r]", " ")
                  echo "haveChangedManifest: ${haveChangedManifest}"
                  if ( (haveChangedManifest.matches( "pkg/manifest/release(.*).go")) || FORCE_RPM_BUILD )  {
                    echo "Found an update to the manifest. Rebuilding the rpm"
                    build job: "../kubekit-rpm-builder/master",
                    parameters: [[$class: 'StringParameterValue', name: 'CLONE_BRANCH', value: "${env.CHANGE_BRANCH}"],], wait: true
                  }
                } else { //Not a PR, build with current branch
                  build job: "../kubekit-rpm-builder/master",
                  parameters: [[$class: 'StringParameterValue', name: 'CLONE_BRANCH', value: "${env.BRANCH}"],], wait: true
                }//if else
              } // Do we need to build a new rpm
              KK_VER = sh(returnStdout: true, script: "awk '\$1 == \"const\" && \$2 == \"Version\" {print \$NF}' ./pkg/manifest/version.go | tr -d '\"'|tr -d '\\n'")
              def (MAJOR, MINOR, PATCH) = KK_VER.split(/\./)
              if ( env.CHANGE_BRANCH && findPlatform( keywords: "(?i)(.*)pkg/manifest/release(.*)") ) {
                kkrpm = findLatestKubekitRpm(ver: env.VERSION, parent: env.GIT_BRANCH )
              } else {
                kkrpm = findLatestKubekitRpm(ver: env.VERSION, parent: "master" )
              }
              KK_RPM_VER = org.apache.commons.lang3.StringUtils.remove(kkrpm,'"')
              KK_RPM_NAME = org.apache.commons.io.FilenameUtils.getName(KK_RPM_VER)
              ARTIFACTORY_PROPERTIES = "kubekit.name=kubekit;kubekit.version=${env.VERSION};kubekit.major=${MAJOR};kubekit.minor=${MINOR};kubekit.patch=${PATCH};kubekit.rpm=${KK_RPM_NAME}"
              getKubekitRpm( OUTPUT_DIR:"./", kubekitRPM: "${KK_RPM_VER}")
              generateChangeLog()
              BuildGoInDockerImage( makeopts: "-j4 build-all" )
              sh(script: 'GOLANG_VER="${GOLANG}" make release-package PKG_BASE=github.com/liferaft')

              def releaseType = "d2d"
              // List all files we want to eventually release to customers here
              if ( BRANCH_NAME.matches("release(.*)") ) { releaseType = "release" }

              def target = "${SNAPSHOT_REPO}/${REPO_PATH}/${VERSION}/${DEV_VERSION}/"
              if ( env.CHANGE_BRANCH ) {target = "${SNAPSHOT_REPO}/${REPO_PATH}/${VERSION}/PR/${BRANCH_NAME}-${DEV_VERSION}/"}

              def uploadSpec = """{
                "files": [
                {
                  "regexp": "true",
                  "pattern": "kubekit_.*linux_amd64.tgz",
                  "props": "${ARTIFACTORY_PROPERTIES};kubekit.release.type=${releaseType};kubekit.arch=amd64;kubekit.os=linux",
                  "target": "${target}"
                },
                {
                  "regexp": "true",
                  "pattern": "kubekit_.*linux_386.tgz",
                  "props": "${ARTIFACTORY_PROPERTIES};kubekit.release.type=${releaseType};kubekit.arch=386;kubekit.os=linux",
                  "target": "${target}"
                },
                {
                  "regexp": "true",
                  "pattern": "kubekit_.*darwin_amd64.tgz",
                  "props": "${ARTIFACTORY_PROPERTIES};kubekit.release.type=${releaseType};kubekit.arch=amd64;kubekit.os=darwin",
                  "target": "${target}"
                },
                {
                  "regexp": "true",
                  "pattern": "kubekit_.*windows_amd64.tgz",
                  "props": "${ARTIFACTORY_PROPERTIES};kubekit.release.type=${releaseType};kubekit.arch=amd64;kubekit.os=windows",
                  "target": "${target}"
                }
                ]
              }"""
              def buildInfo = server.upload(uploadSpec)
              env.BUILD_NAME = buildInfo.name
              server.publishBuildInfo(buildInfo)
              def slackMessage = "Branch: ${BRANCH_NAME} :: https://jfrog.com/artifactory/dependencies-snapshot-sd/${REPO_PATH}/${VERSION}/${DEV_VERSION}/"
              //slackSend ( channel: '#kubekit-cicd', color: 'good', message: "${slackMessage}", tokenCredentialId: "jenkins_slack_credentials" )
            } //script
          } //steps
          post { always { postAction(cleanws: true) } } //post
        } //stage
      } //parallel
    } //stage

    stage('Record Coverage') {
      when { branch 'master' }
      steps {
        script {
          currentBuild.result = 'SUCCESS'
        }
        step([$class: 'MasterCoverageAction', scmVars: [GIT_URL: env.GIT_URL]])
      } //steps
      post { always { postAction(cleanws: true) } } //post
    } //stage

    stage('PR Coverage to Github') {
      when { allOf {not { branch 'master' }; expression { return env.CHANGE_ID != null }} }
      steps {
        script {
          currentBuild.result = 'SUCCESS'
        }
        step([$class: 'CompareCoverageAction', scmVars: [GIT_URL: env.GIT_URL]])
      } //steps
      post { always { postAction(cleanws: true) } } //post
    } //stage

    stage('Parallel') {
      parallel {
        stage ('AKS'){
          agent { node { label DEFAULT_PIPELINE_NODE as String } }
          when {
            expression {
              params.AKS_ENABLED  == true ||
              findPlatform( keywords: "(?i)(.*)aks(.*)" ) ||
              env.BRANCH_NAME.matches( "release(.*)" )
            }
          }
          environment {
            PLAT = "aks"
            MACHINE_NAME = "ak-${env.BASE_MACHINE_NAME}"
          }
          steps {
            getKubekitRpm( OUTPUT_DIR:"./", kubekitRPM: "${env.KK_PATH}/kubekit_*linux_amd64.tgz", explode: "true" )
            platformTest( platformOs:"${PLAT}", numMasters: "1", numWorkers: "3", machineName: "${MACHINE_NAME}" )
          } //steps
          post {
            always { postAction(cleanws: true, platformOs:"${PLAT}", machineName: "${MACHINE_NAME}") }
          } //post
        } //stage
        stage ('Openstack'){
          agent { node { label DEFAULT_PIPELINE_NODE as String } }
          when {
            expression {
              params.OPENSTACK_ENABLED  == true ||
              findPlatform( keywords: "(?i)(.*)openstack(.*)" ) ||
              env.BRANCH_NAME.matches("release(.*)")
            }
          }
          environment {
            PLAT = "openstack"
            MACHINE_NAME = "o-${env.BASE_MACHINE_NAME}"
          }
          steps {
            getKubekitRpm( OUTPUT_DIR:"./", kubekitRPM: "${env.KK_PATH}/kubekit_*linux_amd64.tgz", explode: "true" )
            platformTest( platformOs:"${PLAT}", numMasters: "1", numWorkers: "3", machineName: "${MACHINE_NAME}" )
          } //steps
          post {
            always { postAction(cleanws: true, platformOs:"${PLAT}", machineName: "${MACHINE_NAME}") }
          } //post
        } //stage

        stage ('Openstack Hardened'){
          agent { node { label DEFAULT_BUILD_NODE as String } }
          when { expression {
              params.OPENSTACK_HARDENED_ENABLED == true ||
              findPlatform( keywords: "(?i)(.*)hard(.*)" )
           // || env.BRANCH_NAME.matches("release(.*)")
            }
          }
          environment {
            MACHINE_NAME = "oh-${env.BASE_MACHINE_NAME}"
            PLAT = "openstack"
          }
          steps {
            getKubekitRpm( OUTPUT_DIR:"./", kubekitRPM: "${env.KK_PATH}/kubekit_*linux_amd64.tgz", explode: "true" )
            platformTest( platformOs:"${PLAT}", numMasters: "1", numWorkers: "3", machineName: "${MACHINE_NAME}"  )
          } //steps
          post {
            always { postAction(cleanws: true, platformOs:"${PLAT}", machineName: "${MACHINE_NAME}") }
          } //post
        } //stage

        stage ('Openstack SLES15'){
          agent { node { label DEFAULT_PIPELINE_NODE as String } }
          when { expression {
              params.OPENSTACK_SLES15_ENABLED  == true ||
              findPlatform( keywords: "(?i)(.*)sles15(.*)" ) 
           // || env.BRANCH_NAME.matches("release(.*)")
            }
          }
          environment {
            MACHINE_NAME = "s15-${env.BASE_MACHINE_NAME}"
            PLAT = "openstack"
          }
          steps {
            getKubekitRpm( OUTPUT_DIR:"./", kubekitRPM: "${env.KK_PATH}/kubekit_*linux_amd64.tgz", explode: "true" )
            platformTest( platformOs:"${PLAT}", numMasters: "1", numWorkers: "3", machineName: "${MACHINE_NAME}"  )
          } //steps
          post {
            always { postAction(cleanws: true, platformOs:"${PLAT}", machineName: "${MACHINE_NAME}") }
          } //post
        } //stage

        stage ('EC2'){
          agent { node { label DEFAULT_PIPELINE_NODE as String } }
          when { expression {
              params.EC2_ENABLED  == true ||
              findPlatform( keywords: "(?i)(.*)ec2(.*)" ) ||
              env.BRANCH_NAME.matches( "release(.*)" )
            }
          }
          environment {
            PLAT = "ec2"
            MACHINE_NAME = "a-${env.BASE_MACHINE_NAME}"
          }
          steps {
            getKubekitRpm( OUTPUT_DIR:"./", kubekitRPM: "${env.KK_PATH}/kubekit_*linux_amd64.tgz", explode: "true" )
            platformTest( platformOs:"${PLAT}", numMasters: "1", numWorkers: "3", machineName: "${MACHINE_NAME}"   )
          } //steps
          post {
            always { postAction(cleanws: true, platformOs:"${PLAT}", machineName: "${MACHINE_NAME}") }
          } //post
        } //stage

        stage ('EKS'){
          agent { node { label DEFAULT_PIPELINE_NODE as String } }
          when { expression {
              params.EKS_ENABLED  == true || (
              findPlatform( keywords: "(?i)(.*)eks(.*)" ) ||
              env.BRANCH_NAME.matches("release(.*)") )
            }
          }
          environment {
            PLAT = 'eks'
            MACHINE_NAME = "e-${env.BASE_MACHINE_NAME}"
          }
          steps {
            getKubekitRpm( OUTPUT_DIR:"./", kubekitRPM: "${env.KK_PATH}/kubekit_*linux_amd64.tgz", explode: "true" )
            platformTest( platformOs:"${PLAT}",  machineName: "${MACHINE_NAME}" )
          } //steps
          post {
            always { postAction(cleanws: true, platformOs:"${PLAT}", machineName: "${MACHINE_NAME}") }
          } //post
        } //stage

        stage ('vSphere'){
          agent { node { label DEFAULT_PIPELINE_NODE as String } }
          when {
            expression {
              params.VSPHERE_ENABLED == true ||
              findPlatform( keywords: "(?i)(.*)vsphere(.*)|(.*)vserver(.*)|(.*)vcenter(.*)" ) ||
              env.BRANCH_NAME.matches("master|release(.*)")
            }
          }
          environment {
            PLAT = "vsphere"
            MACHINE_NAME = "v-${env.BASE_MACHINE_NAME}"
          }
          steps {
            script {
            getKubekitRpm( OUTPUT_DIR:"./", kubekitRPM: "${env.KK_PATH}/kubekit_*linux_amd64.tgz", explode: "true" )
            platformTest( platformOs:"${PLAT}", numMasters: "1", numWorkers: "3", machineName: "${MACHINE_NAME}" )
            } //script
          } //steps
          post {
            always { postAction(cleanws: true, platformOs:"${PLAT}", machineName: "${MACHINE_NAME}") }
          } //post
        } //stage
      } //parallel
      post {
        always { archiveLogs() }
        failure { echo "We had some type of failure" }
      }//post
    } //stage Parallel

    stage ('Starting kubekit-release-test job') {
      when { expression { env.BRANCH_NAME.matches("release(.*)|master")  && params.CRON_TRIGGER } }
      options { skipDefaultCheckout() }
      steps {
        build job: "../kubekit-release-test/master",
        parameters: [
          [$class: 'StringParameterValue', name: 'KK_BINARY', value: "${env.KK_PATH}/kubekit_*linux_amd64.tgz"],
        ], wait: false
      } //steps
      post {
        always { postAction(cleanws: true) }
      } //post
    } //stage

    stage ('Promote To QA'){
      agent { node { label DEFAULT_PIPELINE_NODE as String } }
      when { expression { env.BRANCH_NAME.matches("release(.*)|master") && !params.CRON_TRIGGER } }
      options { skipDefaultCheckout() }
      steps {
        echo "Promoting build to QA pool"
        script {
          def promotionConfig = [
            'buildName'           : env.BUILD_NAME,
            'buildNumber'         : env.BUILD_NUMBER,
            'sourceRepo'          : env.SNAPSHOT_REPO,
            'targetRepo'          : env.QA_REPO,
            'status'              : 'QA',
            'comment'             : env.RUN_DISPLAY_URL,
            'includeDependencies' : true,
            'copy'                : true,
            'failFast'            : true
          ]
          server.promote promotionConfig
        }
      }
    }

    stage ('Promote To Stable'){
      agent { node { label DEFAULT_PIPELINE_NODE as String } }
      when { expression { env.BRANCH_NAME.matches("release(.*)") && !params.CRON_TRIGGER } }
      options { skipDefaultCheckout() }
      steps {
        echo "Promoting build to Stable pool"
        script {
          def promotionConfig = [
            'buildName'           : env.BUILD_NAME,
            'buildNumber'         : env.BUILD_NUMBER,
            'sourceRepo'          : env.QA_REPO,
            'targetRepo'          : env.STABLE_REPO,
            'status'              : 'stable',
            'comment'             : env.RUN_DISPLAY_URL,
            'includeDependencies' : true,
            'copy'                : true,
            'failFast'            : true
          ]
          server.promote promotionConfig
        }
      }
    }

    stage ('Decision: Release Build?') {
      agent { node { label DEFAULT_PIPELINE_NODE as String } }
      when { expression { env.BRANCH_NAME.matches("release(.*)") && !params.CRON_TRIGGER } }
      options { skipDefaultCheckout() }
      steps {
        script {
          // Abort older jobs that are still promoting to stable
          milestone 800
          try {
            timeout(time: 1, unit: 'HOURS') {
            env.RELEASE = input message: 'User input required', ok: 'Shall we release the code?',
            parameters: [choice(name: 'RELEASE', choices: 'No\nYes', defaultValue: 'No', description: 'Shall we release the code?')] }
          }
            catch(err) { // timeout reached or input false
              env.RELEASE = 'No'
            } //catch
          // Abort older jobs that are still waiting to release a build
          milestone 900
        }
      }
    }

    stage ('Release Stage'){
      agent { node { label DEFAULT_PIPELINE_NODE as String } }
      when { expression { env.BRANCH_NAME.matches("release(.*)")&& env.RELEASE.matches("Yes") && !params.CRON_TRIGGER } }

      steps {
        echo "Releasing build"
        script {
          currentBuild.rawBuild.keepLog()

          def promotionConfig = [
            'buildName'           : env.BUILD_NAME,
            'buildNumber'         : env.BUILD_NUMBER,
            'sourceRepo'          : env.STABLE_REPO,
            'targetRepo'          : env.RELEASED_REPO,
            'status'              : 'released',
            'comment'             : env.RUN_DISPLAY_URL,
            'includeDependencies' : true,
            'copy'                : true,
            'failFast'            : true
          ]

          server.promote promotionConfig
          currentBuild.rawBuild.keepLog(true)

          slackSend (
            channel: "#kubekit-release",
            color:   'good',
            message: "Release: ${BRANCH_NAME} :: https://jfrog.com/artifactory/dependencies-released-sd/${REPO_PATH}/${VERSION}/${DEV_VERSION}/",
            tokenCredentialId: "jenkins_slack_credentials"
          )

          echo "Tagging the build with git tag -a ${VERSION} -m 'Tag ${BRANCH_NAME} with : ${DEV_VERSION}'"
          withCredentials([
            usernamePassword(
              credentialsId: 'jenkins_github_credentials',
              passwordVariable: 'GIT_PASSWORD',
              usernameVariable: 'GIT_USERNAME'
            )
          ]) {
            sh '''
              git config --global user.name "kubekit.jenkins"
              git config --global user.email kubekit.jenkins@github.com
              gitRepo=$(git config --get remote.origin.url | sed 's#https://##g')
              # tag git commit
              git tag -f -a v${VERSION} -m "Tag ${BRANCH_NAME} with : v${VERSION}"
              git push https://${GIT_USERNAME}:${GIT_PASSWORD}@${gitRepo} --tags
            '''
          } //withCreds
          // Abort older jobs that are still releasing a build
          milestone 1000
        } //script
      } //steps
    } //stage

  } // end stages

  post {
    always {
      script {
        compressBuildLog()
        checkPanic()
        cleanWs()
      }//script
    }//always
  }//post
}
