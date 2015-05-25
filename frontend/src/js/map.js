Handlebars.registerHelper('state', function(options) {
  switch (options.fn(this)) {
    case '0':
      return 'Queued';
    case '1':
      return 'Ready';
    default:
      return 'Other';
  }
});

function start() {
  id = $.urlParam('id');
  if (id) {
    render(URL_B + '/map/' + id, 'map', {});
  } else {
    $('#content').html('No map id given, and I still can\'t read minds.');
  }
}
