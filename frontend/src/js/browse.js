function filter(by) {
  current = $.cookie('filter') || 'by_date';
  if (current == by) {
    return;
  }
  $('#' + current).removeClass('active');
  $('#' + by).addClass('active');
  $.cookie('filter', by);
  refresh();
}

function refresh() {
  current = $.cookie('filter') || 'by_date';
  by = current.slice(3);
  render(URL_B + '/browse/' + by, 'browse');
}

function start() {
  current = $.cookie('filter') || 'by_date';
  $('#' + current).addClass('active');
  refresh();
}
