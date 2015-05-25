function send(url, data, callback) {
  $.ajax({
    url: url,
    method: 'POST',
    data: data,
    crossDomain: true,
    contentType: false,
    processData: false,
    xhrFields: { withCredentials: true },
    beforeSend: function(x, s) {
      s.xhr().upload.addEventListener('progress', function(e) {
        if (e.lengthComputable) {
          percent = (e.loaded / e.total) * 100;
          $('#progressbar').progressbar({value: percent});
        } else {
          $("#progressbar").progressbar({value: false});
        }
      }, false);
    },
    success: function(r, s, x) {
      callback(r);
    },
    error: function(x, s, e) {
      ferror(url, x, s, e);
    },
    complete: function(x, s) {
      $('#progressbar').progressbar('destroy');
    }
  });
}

function upload(doSend) {
  fd = new FormData();
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
  name = atob(id);
  for (var i = 0; i < files.length; i++) {
    m = rex.exec(files[i].name);
    if (!m) {
      $('#umsg').html('Bad file name. Ensure you only add seen files...');
      return;
    }
    if (m[1] != id) {
      oname = atob(m[1]);
      $('#umsg').html('Found characters: ' + [name, oname].join(', ') + '. Please decide...');
      return;
    }
    fd.append('files[]', files[i]);
  }
  $('#umsg').html('Got ' + files.length + ' overmaps for character <b>' + name + '</b>.');
  if (!doSend) {
    return;
  }
  desc = $('#worldname').val();
  if (!desc) {
    $('#umsg').html('You should add some world description...');
    return;
  }
  fd.append('worldname', desc);
  $('#umsg').html('Uploading...');
  $('#progressbar').progressbar({value: 0});
  $('#progressbar').progressbar('enable');
  url = URL_U + '/upload';
  send(url, fd, function(r) {
    if (r.success) {
      $('#umsg').html('Uploaded ' + r.data.uploaded + ' files.');
    } else {
      berror(url, r);
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
