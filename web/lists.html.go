package web

const web_lists_html = `<!doctype html>
<html>

<head>
<title>Lists</title>
<link href="//dn-staticfile.qbox.me/twitter-bootstrap/3.2.0/css/bootstrap.min.css" rel="stylesheet">
<script src="//dn-staticfile.qbox.me/jquery/1.11.3/jquery.min.js"></script>
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
              <th>Download Count</th>
              <th>Category</th>
              <th>Updated At</th>
            </tr>
          </thead>
          <tbody></tbody>
          <tfoot>
            <tr>
              <td colspan="5">
                <div class="btn-group">
                  <button type="button" class="btn btn-default" id="prev">Prev</button>
                  <button type="button" class="btn btn-default" id="next">Next</button>
                </div>
                <select class="form-control" id="pages" style="width: 100px; display: inline-block;"></select>
              </td>
            </tr>
            <tr>
              <td colspan="5">
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
    var getListsWS = new WebSocket('ws://' + window.location.host + '/get-lists');
    getListsWS.onopen = function (evt) {
      $('#get-lists').prop('disabled', false).text('Get Lists');
    };
    getListsWS.onclose = function (evt) {
      $('#get-lists').prop('disabled', true).text('Reload Page to Reconnect');
      getListsWS = null;
    };
    getListsWS.onmessage = function (evt) {
      var resp = JSON.parse(evt.data);
      $('#get-lists').prop('disabled', resp.is_getting_lists).text(resp.status_text);
      get();
    };
    getListsWS.onmessage = function (evt) {
      var resp = JSON.parse(evt.data);
      var text = resp.status_text;
      if (text) text = 'Getting lists: ' + text;
      else      text = 'Get Lists';
      $('#get-lists').prop('disabled', resp.is_getting_lists).text(text);
      get();
    };
    getListsWS.onerror = function (evt) {
      $('#get-lists').prop('disabled', false).text('Get Lists');
    };
    $('#get-lists').on('click', function () {
      if (!getListsWS) return;
      getListsWS.send('get');
      $('#get-lists').prop('disabled', true).text('Loading...');
    });
  </script>
</body>

</html>
`
