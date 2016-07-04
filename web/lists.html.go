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
        <div id="alert" class="alert alert-danger alert-dismissible hidden">
          <button type="button" class="close"><span>&times;</span></button>
          <strong>Error:</strong>
          <span class="text"></span>
        </div>
        <table class="table" id="lists">
          <thead>
            <tr>
              <th><input type="checkbox" name="all"></th>
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
              <td colspan="7">
                <div class="btn-group">
                  <button type="button" class="btn btn-default" id="prev">Prev</button>
                  <button type="button" class="btn btn-default" id="next">Next</button>
                </div>
                <select class="form-control" id="pages" style="width: 100px; display: inline-block;"></select>
                <div class="btn-group pull-right">
                  <button type="button" class="btn btn-default" disabled id="get-lists">Get Lists</button>
                  <button type="button" class="btn btn-default" disabled id="get-dicts">Get Dicts</button>
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
        $('input[name="all"]').prop('checked', false);
        $('#pages').val(currentPage);
        $('#lists tbody').empty();
        $.each(lists, function (_, item) {
          $('#lists tbody').append(
            '<tr>' +
              '<td><input type="checkbox" name="sogou_id" value="' + item.sogou_id + '"></td>' +
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

    var getListsProgressTimeout;

    var getWS = new WebSocket('ws://' + window.location.host + '/get');
    getWS.onopen = function (evt) {
      $('#get-lists, #get-dicts').prop('disabled', false);
    };
    getWS.onclose = function (evt) {
      $('#status').text('Reload Page to Reconnect');
      $('#get-lists, #get-dicts').prop('disabled', true);
      getWS = null;
    };
    getWS.onmessage = function (evt) {
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
        get();
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
    };
    getWS.onerror = function (evt) {
      $('#get-lists, #get-dicts').prop('disabled', false);
    };

    $('#get-lists').on('click', function () {
      if (!getWS) return;
      getWS.send(JSON.stringify({ type: 'get-lists' }));
      $('#get-lists').prop('disabled', true);
    });

    $(document).on('click', '[data-get]', function (e) {
      e.preventDefault();
      if (!getWS) return;
      getWS.send(JSON.stringify({ type: 'get-dicts', value: String($(this).data('get')) }));
    });

    $(document).on('click', '#lists tbody tr', function (e) {
      if (e.target.nodeName === 'INPUT') return;
      var checkbox = $(this).find('input[type="checkbox"]');
      checkbox.prop('checked', !checkbox.prop('checked'));
    });

    $('input[name="all"]').click(function () {
      $('#lists tbody input[name="sogou_id"]').prop('checked', $(this).prop('checked'));
    });

    $('#get-dicts').on('click', function () {
      if (!getWS) return;
      $('input[name="sogou_id"]:checked').each(function (_, item) {
        getWS.send(JSON.stringify({ type: 'get-dicts', value: item.value }));
      });
    });

    $(document).on('click', '.alert-dismissible .close', function () {
      $(this).parent().addClass('hidden');
    });
  </script>
</body>

</html>
`
