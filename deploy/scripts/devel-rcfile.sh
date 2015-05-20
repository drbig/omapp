OMA_DEPLOY_CMDS=(worker uploader backend)

export OMA_DATA_ROOT=$OMA_DEPLOY_ROOT
export OMA_WEB_SECRET=dupa.9
export OMA_DB_DRIVER=sqlite3
export OMA_DB_ARGS=$OMA_DEPLOY_ROOT/db.bin
export OMA_DB_VERBOSE=0
export OMA_WEB_MOUNT=0.0.0.0:7777
export OMA_U_MOUNT=0.0.0.0:8777
export OMA_Q_MOUNT=127.0.0.1:11300

# setup everything at a new root dir
function oma_setup_shop {
  mkdir -p $OMA_DEPLOY_ROOT
  mkdir $OMA_DEPLOY_ROOT/run
  mkdir $OMA_DEPLOY_ROOT/uploads
  mkdir $OMA_DEPLOY_ROOT/maps
  $OMA_DEPLOY_MODEL migrate
  $OMA_DEPLOY_MODEL adduser test test
}

# kill current root dir
function oma_close_shop {
  rm -rf $OMA_DEPLOY_ROOT
  echo "$OMA_DEPLOY_ROOT removed. Note your 'OMA_*' vars are broken now."
  unset OMA_DEPLOY_ROOT
}

# fire up stuff
function oma_fire! {
  if [ -z ${OMA_DEPLOY_RUNNING+x} ]; then
    export OMA_DEPLOY_STAMP=`date +"%Y-%m-%d_%H:%M:%S"`

    echo "Firing beanstalkd..."
    beanstalkd >$OMA_DEPLOY_ROOT/run/beanstalk_${OMA_DEPLOY_STAMP}.log 2>&1 &
    echo $! > $OMA_DEPLOY_ROOT/run/beanstalk_${OMA_DEPLOY_STAMP}.pid

    for CMD in ${OMA_DEPLOY_CMDS[*]}; do
      echo "Firing ${CMD}..."
      $OMA_DEPLOY_PWD/../../cmd/${CMD}/${CMD} >$OMA_DEPLOY_ROOT/run/${CMD}_${OMA_DEPLOY_STAMP}.log 2>&1 &
      echo $! > $OMA_DEPLOY_ROOT/run/${CMD}_${OMA_DEPLOY_STAMP}.pid
    done

    export OMA_DEPLOY_RUNNING=1
    echo "All fired!"
    echo "----------"
    echo "You may like to run 'oma_log_tail'"
  else
    echo "You seem to be running a setup from $OMA_DEPLOY_STAMP"
    echo "Consider oma_retreat!"

  fi
}

# stop and retreat
function oma_retreat! {
  if [ -z ${OMA_DEPLOY_RUNNING+x} ]; then
    echo "You don't seem to be running any setup"
    echo "Consider oma_fire!"
  else
    for CMD in ${OMA_DEPLOY_CMDS[*]}; do
      echo "Retreating ${CMD}..."
      kill `cat $OMA_DEPLOY_ROOT/run/${CMD}_${OMA_DEPLOY_STAMP}.pid`
      rm -f $OMA_DEPLOY_ROOT/run/${CMD}_${OMA_DEPLOY_STAMP}.pid
    done

    # probably should sleep here
    echo "Retreating beanstalkd..."
    kill `cat $OMA_DEPLOY_ROOT/run/beanstalk_${OMA_DEPLOY_STAMP}.pid`
    rm -f $OMA_DEPLOY_ROOT/run/${CMD}_${OMA_DEPLOY_STAMP}.pid

    unset OMA_DEPLOY_RUNNING
    echo "All retreated!"
    echo "--------------"
    echo "You may like to run 'oma_log_last' to list that session logs"
  fi
}

# tail current logs
function oma_log_tail {
  if [ -z ${OMA_DEPLOY_RUNNING+x} ]; then
    echo "You don't seem to be running any setup"
    echo "Consider oma_fire!"
  else
    tail -f $OMA_DEPLOY_ROOT/run/*_${OMA_DEPLOY_STAMP}.log
  fi
}

# list last logs
function oma_log_last {
  if [ -z ${OMA_DEPLOY_STAMP+x} ]; then
    echo "You don't seem to have any last logs"
  else
    ls -1 $OMA_DEPLOY_ROOT/run/*_${OMA_DEPLOY_STAMP}.log
  fi
}

if [ -d $OMA_DEPLOY_ROOT ]; then
  echo "$OMA_DEPLOY_ROOT exists, assuming properly setup."
else
  echo "Setting up shop at $OMA_DEPLOY_ROOT..."
  oma_setup_shop
fi

if [ -e ~/.bashrc ]; then
  source ~/.bashrc
fi

echo "Entering prepped bash instance..."
echo "---------------------------------"
echo "Enable DB debug with 'export OMA_DB_VERBOSE=1'"
echo "Look for useful commands starting with 'oma_'"
