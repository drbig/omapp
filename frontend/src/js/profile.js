function start() {
  user = $.urlParam('user') || $.cookie('user');
  if (user) {
    render(URL_B + '/user/' + user + '/info', 'profile', {user: user});
  } else {
    get(URL_B + '/user', function(r) {
      if (r.success) {
        $.cookie("user", r.data);
        render(URL_B + '/user/' + r.data + '/info', 'profile', {user: r.data});
      } else {
        $('#content').html(Handlebars.partials['logreg.hbs']({target: 'profile', clear: false}));
      }
    });
  }
}
