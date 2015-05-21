var URL_B = 'http://127.0.0.1:7777'
var URL_U = 'http://127.0.0.1:8777'

function request(url, template, args) {
  $.ajax({
    url: url,
    crossDomain: true,
    success: function(r, s, x) {
      r = $.extend(r, args);
      $('#content').html(Handlebars.partials[template + '.hbs'](r));
    },
    error: function(x, s, e) {
      msg = ['All I can tell:', x.statusText, s, e].join(', ') + '.';
      r = $.extend({msg: msg}, args);
      $('#content').html(Handlebars.partials['error.hbs'](r));
    }
  });
}

function start() {
  user = window.location.hash.slice(1);
  if (user == "") {
    user = 'test';
  }
  request(URL_B + '/user/' + user + '/info', 'profile', {user: user});
}
