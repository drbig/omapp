function upload() {
  fd = new FormData();
  name = $('#worldname').val();
  if (!name) {
    $('#umsg').html('You should add some world description...');
    return;
  }
  fd.append('worldname', name);
  files = $('#files')[0].files;
  if (files.length < 1) {
    $('#umsg').html('You need to select some seen files...');
    return;
  }
  rex = new RegExp("\#(.*?)\.seen");
  m = rex.exec(files[0].name);
  if (!m) {
    $('#umsg').html('Bad file name. Ensure you only add seen files...');
    return;
  }
  id = m[1];
  for (var i = 1; i < files.length; i++) {
    m = rex.exec(files[i].name);
    if (!m) {
      $('#umsg').html('Bad file name. Ensure you only add seen files...');
      return;
    }
    if (m[1] != id) {
      $('#umsg').html('Found another character\'s seen file. Please decide...');
      return;
    }
    fd.append('files[]', files[i]);
  }
  $('#umsg').html('Uploading...');
  url = URL_U + '/upload';
  send(url, fd, function(r) {
    if (r.success) {
      $('#umsg').html('Uploaded!');
    } else {
      r = {msg: r.data, url: url, src: 'Backend', ver: r.ver};
      $('#content').html(Handlebars.partials['error.hbs'](r));
    }
  });
}

function start() {
  user = $.cookie('user');
  if (!user) {
    get(URL_B + '/user', function(r) {
      if (r.success) {
        $.cookie("user", r.data);
      } else {
        $('#content').html(Handlebars.partials['logreg.hbs']({target: null, clear: true}));
      }
    });
  }
}
