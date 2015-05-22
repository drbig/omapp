$.urlParam = function(name){
  var results = new RegExp('[\?&]' + name + '=([^&#]*)').exec(window.location.href);
  if (results == null) {
    return null;
  } else {
    return results[1] || 0;
  }
}

Handlebars.registerHelper('times', function(n, block) {
  accum = '';
  for (i = 0; i < n; ++i)
    accum += block.fn(i);
  return accum;
});

Handlebars.registerHelper('date', function(options) {
  d = new Date(options.fn(this));
  return d.toUTCString();
});

function ferror(url, x, s, e) {
  msg = [x.statusText, s, e].join(', ') + '.';
  r = {msg: msg, url: url, src: 'Frontend', ver: 'unknown'};
  $('#content').html(Handlebars.partials['error.hbs'](r));
}

function berror(url, r) {
  r = {msg: r.data, url: url, src: 'Backend', ver: r.ver};
  $('#content').html(Handlebars.partials['error.hbs'](r));
}

function fetch(url, method, data, callback) {
  $.ajax({
    url: url,
    method: method,
    data: data,
    crossDomain: true,
    xhrFields: { withCredentials: true },
    success: function(r, s, x) {
      callback(r);
    },
    error: function(x, s, e) {
      ferror(url, x, s, e);
    }
  });
}

function get(url, callback) {
  fetch(url, 'GET', null, callback);
}

function post(url, data, callback) {
  fetch(url, 'POST', data, callback);
}

function render(url, template, args) {
  get(url, function(r) {
    if (r.success) {
      r = $.extend(r, args);
      $('#content').html(Handlebars.partials[template + '.hbs'](r));
    } else {
      berror(url, r);
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
      berror(url, r);
    }
  });
}
