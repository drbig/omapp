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
  page = $.urlParam('page') || 1;
  current = $.cookie('filter') || 'by_date';
  by = current.slice(3);
  url = URL_B + '/browse/' + by + '?page=' + page;
  get(url, function(r) {
    if (r.success) {
      paginator = '<div class="text-center"><ul class="pagination">';
      for (i = 1; i <= r.data.pages; i++) {
        if (i == r.data.page) {
          paginator += '<li><a class="active" href="?page=' + i + '">'+i+'</a></li>';
        } else {
          paginator += '<li><a href="?page=' + i + '">'+i+'</a></li>';
        }
      }
      paginator += '</ul></div>';
      r = $.extend({paginator: paginator}, r);
      $('#content').html(Handlebars.partials['browse.hbs'](r));
    } else {
      berror(url, r);
    }
  });
}

function start() {
  current = $.cookie('filter') || 'by_date';
  $('#' + current).addClass('active');
  refresh();
}
