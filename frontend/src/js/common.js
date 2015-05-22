$.urlParam = function(name){
  var results = new RegExp('[\?&]' + name + '=([^&#]*)').exec(window.location.href);
  if (results == null) {
    return null;
  } else {
    return results[1] || 0;
  }
}

Handlebars.registerHelper('date', function(options) {
  d = new Date(options.fn(this));
  return d.toUTCString();
});

function fetch(url, method, data, enc, pd, callback) {
  $.ajax({
    url: url,
    method: method,
    data: data,
    contentType: enc,
    crossDomain: true,
    processData: pd,
    xhrFields: { withCredentials: true },
    success: function(r, s, x) {
      callback(r);
    },
    error: function(x, s, e) {
      msg = ['All I can tell:', x.statusText, s, e].join(', ') + '.';
      r = {msg: msg, url: url, src: 'Frontend', ver: 'unknown'};
      $('#content').html(Handlebars.partials['error.hbs'](r));
    }
  });
}

function get(url, callback) {
  fetch(url, 'GET', null, 'application/x-www-form-urlencoded; charset=UTF-8', true, callback);
}

function post(url, data, callback) {
  fetch(url, 'POST', data, 'application/x-www-form-urlencoded; charset=UTF-8', true, callback);
}

function send(url, data, callback) {
  fetch(url, 'POST', data, false, false, callback);
}

function render(url, template, args) {
  get(url, function(r) {
    if (r.success) {
      r = $.extend(r, args);
      $('#content').html(Handlebars.partials[template + '.hbs'](r));
    } else {
      r = {msg: r.data, url: url, src: 'Backend', ver: r.ver};
      $('#content').html(Handlebars.partials['error.hbs'](r));
    }
  });
}

function logreg(template, clear) {
  user = $('#login').val();
  if (user == "") {
    $('#msg').html("You need to enter a login...");
    return
  }
  pass = $('#password').val();
  if (pass == "") {
    $('#msg').html("You need to enter a password...");
    return
  }
  url = URL_B + '/user/' + user;
  post(url, {password: pass}, function(r) {
    if (r.success) {
      $.cookie("user", user);
      if (template) {
        render(URL_B + '/user/' + user + '/info', template, {user: user});
      } else if (clear) {
        $('#content').html('');
      }
    } else {
      r = {msg: r.data, url: url, src: 'Backend', ver: r.ver};
      $('#content').html(Handlebars.partials['error.hbs'](r));
    }
  });
}
