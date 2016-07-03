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
        <table class="table" id="lists">
          <thead>
            <tr>
              <th>ID</th>
              <th>Name</th>
              <th>Suggestions</th>
              <th>Downloads</th>
              <th>Category</th>
              <th>Updated At</th>
            </tr>
          </thead>
          <tbody></tbody>
          <tfoot>
            <tr>
              <td colspan="6">
                <div class="btn-group">
                  <button type="button" class="btn btn-default" id="prev">Prev</button>
                  <button type="button" class="btn btn-default" id="next">Next</button>
                </div>
                <select class="form-control" id="pages" style="width: 100px; display: inline-block;"></select>
              </td>
            </tr>
            <tr>
              <td colspan="6">
                <div class="btn-group">
                  <button type="button" class="btn btn-default" disabled id="get-lists">Get Lists</button>
                </div>
              </td>
            </tr>
          </tfoot>
        </table>
      </div>
    </div>
  </div>
  <script>
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
    var itemsPerPage = 10;
    var currentPage = 1;
    var getting = false;
    function get () {
      if (getting) return;
      getting = true;
      $.getJSON('/lists', { per: itemsPerPage, page: currentPage }).then(function (lists, _, xhr) {
        var totalItems = +xhr.getResponseHeader('Total-Items');
        $('#total').html(totalItems.toLocaleString());
        $('#pages').empty();
        var pages = Math.ceil(totalItems / itemsPerPage);
        for (var i = 0; i < pages; i++) {
          $('#pages').append('<option value="' + (i + 1) + '">' + (i + 1) + '</option>');
        }
        $('#prev, #next, #pages').prop('disabled', pages < 2);
        $('#pages').val(currentPage);
        $('#lists tbody').empty();
        $.each(lists, function (_, item) {
          $('#lists tbody').append(
            '<tr>' +
              '<td>' + item.id + '</td>' +
              '<td><a href="http://pinyin.sogou.com/dict/detail/index/' + item.sogou_id + '" target="_blank">' + decode(item.name) + '</a></td>' +
              '<td><a href data-get="' + item.sogou_id + '">' + item.suggestion_count.toLocaleString() + '</a></td>' +
              '<td>' + item.download_count.toLocaleString() + '</td>' +
              '<td>' + decode(item.category_name) + '</td>' +
              '<td>' + format(item.updated_at) + '</td>' +
            '</tr>'
          );
        });
      }).always(function () {
        getting = false;
      });
    }
    get();
    $('#pages').on('change', function () {
      currentPage = $('#pages').val();
      get();
    });
    $('#prev').on('click', function () {
      if (currentPage > 1) {
        currentPage--;
        $('#pages').val(currentPage).trigger('change');
      }
    });
    $('#next').on('click', function () {
      currentPage++;
      $('#pages').val(currentPage).trigger('change');
    });

    var getWS = new WebSocket('ws://' + window.location.host + '/get');
    getWS.onopen = function (evt) {
      $('#get-lists').prop('disabled', false).text('Get Lists');
    };
    getWS.onclose = function (evt) {
      $('#get-lists').prop('disabled', true).text('Reload Page to Reconnect');
      getWS = null;
    };
    getWS.onmessage = function (evt) {
      var resp = JSON.parse(evt.data);
      if (resp.error) {
        alert(resp.error);
        return;
      }
      switch (resp.type) {
      case 'get-lists':
        var text = resp.status_text;
        if (text) text = 'Getting lists: ' + text;
        else      text = 'Get Lists';
        $('#get-lists').prop('disabled', resp.is_getting_lists).text(text);
        get();
        break;
      case 'get-dicts':
        var id = resp.value;
        var link = $('a[data-get="' + id + '"]');
        if (link.length) {
          link.replaceWith('<div class="progress" data-get="' + id + '"><div class="progress-bar" style="width: 0%;">0%</div></div>');
        }
        var percent = resp.done / resp.total * 100;
        percent = Math.max(+percent.toFixed(2), 0) + '%';
        $('div[data-get="' + id + '"] .progress-bar').text(percent).css('width', percent);
        if (resp.period === 'imported' && resp.done === resp.total) {
          $('[data-get="' + id + '"]').replaceWith('<a href data-get="' + id + '">' + resp.total.toLocaleString() + '</a>');
        }
        break;
      }
    };
    getWS.onerror = function (evt) {
      $('#get-lists').prop('disabled', false).text('Get Lists');
    };

    $('#get-lists').on('click', function () {
      if (!getWS) return;
      getWS.send(JSON.stringify({ type: 'get-lists' }));
      $('#get-lists').prop('disabled', true).text('Loading...');
    });

    $(document).on('click', '[data-get]', function (e) {
      e.preventDefault();
      if (!getWS) return;
      getWS.send(JSON.stringify({ type: 'get-dicts', value: String($(this).data('get')) }));
    });
  </script>
</body>

</html>
`
