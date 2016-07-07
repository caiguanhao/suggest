package web

const web_index_html = `<!doctype html>
<html>

<head>
<title>Suggest</title>
<link href="//dn-staticfile.qbox.me/twitter-bootstrap/3.2.0/css/bootstrap.min.css" rel="stylesheet">
<script src="//dn-staticfile.qbox.me/jquery/1.11.3/jquery.min.js"></script>
<script src="//dn-staticfile.qbox.me/typeahead.js/0.11.1/typeahead.bundle.min.js"></script>
<style>
  /* https://github.com/bassjobsen/typeahead.js-bootstrap-css */
  span.twitter-typeahead .tt-menu,
  span.twitter-typeahead .tt-dropdown-menu {
    position: absolute;
    top: 100%;
    left: 0;
    z-index: 1000;
    display: none;
    float: left;
    min-width: 160px;
    padding: 5px 0;
    margin: 2px 0 0;
    list-style: none;
    font-size: 14px;
    text-align: left;
    background-color: #ffffff;
    border: 1px solid #cccccc;
    border: 1px solid rgba(0, 0, 0, 0.15);
    border-radius: 4px;
    -webkit-box-shadow: 0 6px 12px rgba(0, 0, 0, 0.175);
    box-shadow: 0 6px 12px rgba(0, 0, 0, 0.175);
    background-clip: padding-box;
    width: 100%;
  }
  span.twitter-typeahead .tt-item,
  span.twitter-typeahead .tt-suggestion {
    display: block;
    padding: 3px 20px;
    clear: both;
    font-weight: normal;
    line-height: 1.42857143;
    color: #333333;
    white-space: nowrap;
  }
  span.twitter-typeahead .tt-suggestion.tt-cursor,
  span.twitter-typeahead .tt-suggestion:hover,
  span.twitter-typeahead .tt-suggestion:focus {
    color: #ffffff;
    text-decoration: none;
    outline: 0;
    background-color: #337ab7;
  }
  span.twitter-typeahead {
    width: 100%;
  }
  .hl {
    color: #cb1c00;
  }
  span.twitter-typeahead .tt-suggestion.tt-cursor .hl,
  span.twitter-typeahead .tt-suggestion:hover .hl,
  span.twitter-typeahead .tt-suggestion:focus .hl {
    color: #ffffff;
  }
</style>
</head>

<body>
  <div class="container">
    <div class="row">
      <div class="col-md-offset-1 col-md-10 col-sm-12 col-xs-12">
        <div class="page-header">
          Suggest (<span id="total">0</span>)
          /
          <a href="/lists">Lists</a>
        </div>
        <input type="text" class="form-control" placeholder="Search for..." id="search">
        <hr>
        <div class="btn-toolbar">
          <div class="btn-group form-control-static">
            <strong>Using:</strong>
          </div>
          <div class="btn-group">
            <button type="button" class="btn btn-default" data-using="ws">WebSocket</button>
            <button type="button" class="btn btn-default" data-using="http">HTTP</button>
          </div>
        </div>
      </div>
    </div>
  </div>

  <script>
    (function () {
      var getWS;
      var current = {};
      var msWaited = 0;

      var sourceHTTP = new Bloodhound({
        datumTokenizer: Bloodhound.tokenizers.whitespace,
        queryTokenizer: Bloodhound.tokenizers.whitespace,
        remote: {
          url: '/suggestions?q=_QUERY_',
          wildcard: '_QUERY_',
          rateLimitWait: 0,
          transport: function (settings, success, error) {
            var startTime = +new Date();
            $.ajax(settings).fail(error).done(success).always(function () {
              msWaited = +new Date() - startTime;
            });
          }
        }
      });

      var sourceWS = new Bloodhound({
        datumTokenizer: Bloodhound.tokenizers.whitespace,
        queryTokenizer: Bloodhound.tokenizers.whitespace,
        remote: {
          url: '_QUERY_',
          wildcard: '_QUERY_',
          transport: function (settings, success, error) {
            if (!getWS) return;
            var startTime = +new Date();
            getWS.send(JSON.stringify({ type: 'get-suggestions', value: settings.url }));
            getWS.onmessage = function (evt) {
              msWaited = +new Date() - startTime;
              var resp = JSON.parse(evt.data);
              if (resp.error) {
                error(resp);
                return;
              }
              success(resp);
            };
          },
          rateLimitWait: 0
        }
      });

      function setCurrent () {
        var a = window.location.search.substr(1).split('&');
        var qs = {};
        for (var i = 0; i < a.length; i++) {
          var p = a[i].split('=');
          if (p.length !== 2) continue;
          qs[p[0]] = decodeURIComponent(p[1].replace(/\+/g, " "));
        }
        current.using = qs.using || undefined;
        for (var key in current) {
          if (!current[key]) delete current[key];
        }
        var using = current.using;
        if (using !== 'http') using = 'ws';
        $('[data-using]').removeClass('active');
        $('[data-using="' + using + '"]').addClass('active');
      }

      function getSuggestionCount () {
        $.getJSON('/approximate_suggestion_count').then(function (count) {
          $('#total').html(count.approximate_suggestion_count.toLocaleString());
        });
      }

      function setupWebSocket () {
        getWS = new WebSocket('ws://' + window.location.host + '/get');
        getWS.onopen = function (evt) {};
        getWS.onclose = function (evt) {
          getWS = null;
        };
        getWS.onerror = function (evt) {};
      }

      function setupControls () {
        $(document).on('click', '[data-using]', function () {
          $('[data-using]').not(this).removeClass('active');
          $(this).addClass('active');
          $('#search').typeahead('destroy');
          current.using = $(this).data('using');
          setupTypeAhead();
        });
        $(window).on('popstate', function (event) {
          setCurrent();
        });
      }

      function setupTypeAhead () {
        var source;
        if (current.using === 'http') {
          source = sourceHTTP;
        } else {
          source = sourceWS;
        }
        $('#search').typeahead(null, {
          limit: 20,
          source: source,
          display: 'text',
          templates: {
            suggestion: function (sugg) {
              var text = sugg.text;
              if (sugg.start > -1) {
                text = text.slice(0, sugg.start) +
                  '<span class="hl">' + text.slice(sugg.start, sugg.end) + '</span>' +
                  text.slice(sugg.end);
              }
              return '<div>' + text + '</div>';
            },
            notFound: function () {
              return '<div class="tt-item">Sorry, no suggestion for you.</div>';
            },
            footer: function () {
              var footer = '<div class="tt-item"><small class="text-muted">';
              if (msWaited > 0) {
                footer += 'Search took ' + msWaited + ' ms.';
                msWaited = 0;
              } else {
                footer += 'Cached results shown.';
              }
              footer += '</small></div>';
              return footer;
            }
          }
        });

        var c = $.extend(true, {}, current);
        if (c.using === 'ws') delete c.using;
        var qs = $.param(c);
        if (qs) qs = '?' + qs;
        if (window.location.search !== qs) {
          window.history.pushState({}, '', window.location.pathname + qs);
        }
      }

      function main () {
        setCurrent();
        getSuggestionCount();
        setupWebSocket();
        setupControls();
        setupTypeAhead();
      }

      main();
    })();
  </script>
</body>

</html>
`
