function start() {
  id = $.urlParam('id');
  if (id) {
    render(URL_B + '/map/' + id, 'map', {});
  } else {
    $('#content').html('No map id given, and I still can\'t read minds.');
  }
}
