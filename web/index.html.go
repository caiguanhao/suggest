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
</style>
</head>

<body>
  <div class="container">
    <div class="row">
      <div class="col-md-offset-1 col-md-10 col-sm-12 col-xs-12">
        <div class="page-header">
          Suggest
          /
          <a href="/lists">Lists</a>
        </div>
        <input type="text" class="form-control" placeholder="Search for..." id="search">
      </div>
    </div>
  </div>

  <script>
    $('#search').typeahead(null, {
      limit: 20,
      source: new Bloodhound({
        datumTokenizer: Bloodhound.tokenizers.whitespace,
        queryTokenizer: Bloodhound.tokenizers.whitespace,
        remote: {
          url: '/suggestions?q=_QUERY_',
          wildcard: '_QUERY_',
          rateLimitWait: 0
        }
      }),
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
        }
      }
    });
  </script>
</body>

</html>
`
