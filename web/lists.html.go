package web

const web_lists_html = `<!doctype html>
<html>

<head>
<title>Lists</title>
<link href="//dn-staticfile.qbox.me/twitter-bootstrap/3.2.0/css/bootstrap.min.css" rel="stylesheet">
<script src="//dn-staticfile.qbox.me/jquery/1.11.3/jquery.min.js"></script>
<style>
.progress {
  margin-bottom: 0;
}
[data-sortable] {
  cursor: pointer;
}
[data-sortable]:after {
  content: '▬';
  float: right;
  color: #666;
  font-size: 10px;
  line-height: 20px;
}
.sorted[data-sortable]:after {
  content: '▲';
}
.sorted.desc[data-sortable]:after {
  content: '▼';
}
</style>
</head>

<body>
  <div class="container">
    <div class="row">
      <div class="col-md-offset-1 col-md-10 col-sm-12 col-xs-12">
        <div class="page-header">
          <a href="/">Suggest</a>
          /
          Lists (<span id="total">0</span>)
        </div>
        <div id="alert" class="alert alert-danger alert-dismissible hidden">
          <button type="button" class="close"><span>&times;</span></button>
          <strong>Error:</strong>
          <span class="text"></span>
        </div>
        <table class="table" id="lists">
          <thead>
            <tr>
              <th><input type="checkbox" name="all"></th>
              <th data-sortable="sogou_id">ID</th>
              <th data-sortable="name">Name</th>
              <th data-sortable="suggestion">Suggestions</th>
              <th data-sortable="download">Downloads</th>
              <th data-sortable="category">Category</th>
              <th data-sortable="updated_at">Updated At</th>
            </tr>
          </thead>
          <tbody></tbody>
          <tfoot>
            <tr>
              <td colspan="7">
                <div class="btn-toolbar">
                  <div class="btn-group">
                    <input type="text" id="search" class="form-control" placeholder="Search for...">
                  </div>
                  <div class="btn-group">
                    <button type="button" class="btn btn-default" id="prev">Prev</button>
                    <button type="button" class="btn btn-default" id="next">Next</button>
                  </div>
                  <div class="btn-group">
                    <select class="form-control" id="pages"></select>
                  </div>
                  <div class="btn-group pull-right">
                    <button type="button" class="btn btn-default" disabled id="get-lists">Get Lists</button>
                    <button type="button" class="btn btn-default" disabled id="get-dicts">Get Dicts</button>
                  </div>
                </div>
              </td>
            </tr>
            <tr id="lists-progress" class="hidden">
              <td colspan="7">
                <div class="progress">
                  <div class="progress-bar progress-bar-success"></div>
                </div>
              </td>
            </tr>
            <tr>
              <td colspan="7">
                <div id="status">Ready.</div>
              </td>
            </tr>
          </tfoot>
        </table>
      </div>
    </div>
  </div>
  <script>
    (function () {
      var itemsPerPage = 10;
      var current = {};
      var totalPages = 0;
      var getting = false;
      var getWS = null;
      var getListsProgressTimeout;

      function decode (base64) {
        return decodeURIComponent(escape(atob(base64)));
      }

      function pad (n) {
        return n < 10 ? '0' + n : n;
      }

      function format (d) {
        d = new Date(d);
        return d.getFullYear() + '-' + pad(d.getMonth() + 1) + '-' + pad(d.getDate()) + ' ' +
          pad(d.getHours()) + ':'  + pad(d.getMinutes()) + ':' + pad(d.getSeconds());
      }

      function setCurrent () {
        var a = window.location.search.substr(1).split('&');
        var qs = {};
        for (var i = 0; i < a.length; i++) {
          var p = a[i].split('=');
          if (p.length !== 2) continue;
          qs[p[0]] = decodeURIComponent(p[1].replace(/\+/g, " "));
        }
        current.page = Math.max(+qs.page || 1, 1);
        current.query = qs.query || undefined;
        current.category = +qs.category || undefined;
        current.order = qs.order || undefined;
        for (var key in current) {
          if (!current[key]) delete current[key];
        }
        var o = current.order;
        if (!o) o = '-download';
        var isDESC = o.slice(0, 1) === '-';
        if (isDESC) o = o.slice(1);
        $('[data-sortable="' + o + '"]').addClass('sorted')[isDESC ? 'addClass' : 'removeClass']('desc');
      }

      function getLists () {
        if (getting) return;
        getting = true;

        $.getJSON('/lists', {
          per: itemsPerPage,
          page: current.page,
          q: current.query,
          category_id: current.category,
          order: current.order
        }).then(function (lists, _, xhr) {
          var totalItems = +xhr.getResponseHeader('Total-Items');
          $('#total').html(totalItems.toLocaleString());
          $('#pages').empty();
          totalPages = Math.ceil(totalItems / itemsPerPage);
          for (var i = 0; i < totalPages; i++) {
            $('#pages').append('<option value="' + (i + 1) + '">' + (i + 1) + '</option>');
          }
          $('#prev, #next, #pages').prop('disabled', totalPages < 2);
          $('input[name="all"]').prop('checked', false);
          $('#pages').val(current.page);
          $('#search').val(current.query);
          $('#lists tbody').empty();
          $.each(lists, function (_, item) {
            $('#lists tbody').append(
              '<tr>' +
                '<td><input type="checkbox" name="sogou_id" value="' + item.sogou_id + '"></td>' +
                '<td>' + item.sogou_id + '</td>' +
                '<td><a href="http://pinyin.sogou.com/dict/detail/index/' + item.sogou_id + '" target="_blank">' + decode(item.name) + '</a></td>' +
                '<td><a href data-get="' + item.sogou_id + '">' + item.suggestion_count.toLocaleString() + '</a></td>' +
                '<td>' + item.download_count.toLocaleString() + '</td>' +
                '<td><a href data-category="' + item.category_id + '">' + decode(item.category_name) + '</a></td>' +
                '<td>' + format(item.updated_at) + '</td>' +
              '</tr>'
            );
          });

          var c = $.extend(true, {}, current);
          if (c.page === 1) delete c.page;
          var qs = $.param(c);
          if (qs) qs = '?' + qs;
          if (window.location.search !== qs) {
            window.history.pushState({}, '', window.location.pathname + qs);
          }
        }).always(function () {
          getting = false;
        });
      }

      function setupPaginator () {
        $('#pages').on('change', function () {
          current.page = +$(this).val();
          getLists();
        });
        $('#prev').on('click', function () {
          if (current.page > 1) {
            current.page--;
            getLists();
          }
        });
        $('#next').on('click', function () {
          if (current.page + 1 <= totalPages) {
            current.page++;
            getLists();
          }
        });
        $(window).on('popstate', function (event) {
          setCurrent();
          getLists();
        });
      }

      function setupFilter () {
        $('#search').on('keyup', function (e) {
          var query = $(this).val();
          if (e.which === 13) {
            current.query = query || undefined;
            current.page = 1;
            getLists();
          }
        });
        $(document).on('click', '[data-category]', function (e) {
          e.preventDefault();
          current.category = $(this).data('category');
          current.page = 1;
          getLists();
        });
      }

      function setupSorter () {
        $(document).on('click', '[data-sortable]', function () {
          $('[data-sortable]').not(this).removeClass('sorted');
          $(this).addClass('sorted').toggleClass('desc');
          var dir = $(this).hasClass('desc') ? '-' : '';
          current.order = dir + $(this).data('sortable');
          getLists();
        });
      }

      function onMessage (evt) {
        var resp = JSON.parse(evt.data);
        if (resp.error) {
          $('#alert').removeClass('hidden').find('.text').text(resp.error);
          return;
        }
        switch (resp.type) {
        case 'get-lists-progress':
          var percent = Math.floor(resp.done / resp.total * 100) + '%';
          $('#lists-progress').removeClass('hidden');
          var status = resp.done.toLocaleString() + ' / ' + resp.total.toLocaleString() + ' - ' + percent;
          $('#lists-progress .progress-bar').text(status).css('width', percent);
          $('#get-lists').prop('disabled', true);
          if (resp.done === resp.total) {
            if (getListsProgressTimeout) {
              clearTimeout(getListsProgressTimeout);
            }
            getListsProgressTimeout = setTimeout(function () {
              $('#lists-progress').addClass('hidden');
              $('#lists-progress .progress-bar').text('').css('width', '0%');
              $('#get-lists').prop('disabled', false);
              $('#status').text('');
              getListsProgressTimeout = undefined;
            }, 3000);
          }
          break;
        case 'get-lists':
          $('#status').text(resp.status_text);
          getLists();
          break;
        case 'get-dicts-progress':
          var id = resp.value;
          var link = $('a[data-get="' + id + '"]');
          if (link.length) {
            link.replaceWith('<div class="progress" data-get="' + id + '">' +
              '<div class="progress-bar progress-bar-success" style="width: 0%;">0%</div></div>');
          }
          var percent = resp.done / resp.total * 100;
          percent = Math.max(+percent.toFixed(2), 0) + '%';
          $('div[data-get="' + id + '"] .progress-bar').text(percent).css('width', percent);
          break;
        case 'get-dicts-done':
          var id = resp.value;
          $('[data-get="' + id + '"]').replaceWith('<a href data-get="' + id + '">' + resp.total.toLocaleString() + '</a>');
          break;
        }
      }

      function setupWebSocket () {
        getWS = new WebSocket('ws://' + window.location.host + '/get');
        getWS.onopen = function (evt) {
          $('#get-lists, #get-dicts').prop('disabled', false);
        };
        getWS.onclose = function (evt) {
          $('#status').text('Reload Page to Reconnect');
          $('#get-lists, #get-dicts').prop('disabled', true);
          getWS = null;
        };
        getWS.onmessage = onMessage;
        getWS.onerror = function (evt) {
          $('#get-lists, #get-dicts').prop('disabled', false);
        };
      }

      function setupWebsocketControls () {
        $(document).
        on('click', '#lists tbody tr', function (e) {
          if (e.target.nodeName === 'INPUT') return;
          var checkbox = $(this).find('input[type="checkbox"]');
          checkbox.prop('checked', !checkbox.prop('checked'));
        }).
        on('click', 'input[name="all"]', function () {
          $('#lists tbody input[name="sogou_id"]').prop('checked', $(this).prop('checked'));
        }).
        on('click', '#get-lists', function () {
          if (!getWS) return;
          getWS.send(JSON.stringify({ type: 'get-lists' }));
          $('#get-lists').prop('disabled', true);
        }).
        on('click', '[data-get]', function (e) {
          e.preventDefault();
          if (!getWS) return;
          getWS.send(JSON.stringify({ type: 'get-dicts', value: String($(this).data('get')) }));
        }).
        on('click', '#get-dicts', function () {
          if (!getWS) return;
          $('input[name="sogou_id"]:checked').each(function (_, item) {
            getWS.send(JSON.stringify({ type: 'get-dicts', value: item.value }));
          });
        }).
        on('click', '.alert-dismissible .close', function () {
          $(this).parent().addClass('hidden');
        });
      }

      function main () {
        setCurrent();
        getLists();
        setupPaginator();
        setupFilter();
        setupSorter();
        setupWebSocket();
        setupWebsocketControls();
      }

      main();
    })();
  </script>
</body>

</html>
`
