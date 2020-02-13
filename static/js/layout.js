// FAB toggle
function toggleFAB() {
  var fabContainer = document.querySelector('.fab-message')
  var fabButton = fabContainer.querySelector('.fab-message-button a')
  var fabToggle = document.getElementById('fab-message-toggle')
  fabContainer.classList.toggle('is-open')
  fabButton.classList.toggle('toggle-icon')
}
$(document).ready(function() {
  var fabContainer = document.querySelector('.fab-message')
  var messages = document.querySelector('.fab-message-content h3')
  if (messages) {
    fabContainer.style.display = 'initial'
  }
})

// Theme switch
function switchTheme(e) {
  if (e.target.checked) {
    document.documentElement.setAttribute('data-theme', 'light')
    localStorage.setItem('theme', 'light')
    document.getElementById('nav').classList.remove('navbar-dark')
    document.getElementById('nav').classList.add('navbar-light')
  } else {
    document.documentElement.setAttribute('data-theme', 'dark')
    document.getElementById('nav').classList.remove('navbar-light')
    document.getElementById('nav').classList.add('navbar-dark')
    localStorage.setItem('theme', 'dark')
  }
}
$('#toggleSwitch').on('change', switchTheme)


moment.locale((window.navigator.userLanguage || window.navigator.language).toLowerCase())
$('[aria-local-date]').each(function(item) {
  var dt = $(this).attr('aria-local-date')
  var format = $(this).attr('aria-local-date-format')

  if (!format) {
    format = 'L LTS'
  }

  if (format === 'FROMNOW') {
    $(this).text(moment.unix(dt).fromNow())
  } else {
    $(this).text(moment.unix(dt).format(format))
  }
})

var indicator = $('#nav .nav-indicator')
var items = document.querySelectorAll('#nav .nav-item')
var selectedLi = indicator.parent()[0]
var navigated = false

function handleIndicator(el) {
  indicator.css({
    width: `${el.offsetWidth}px`,
    left: `${el.offsetLeft}px`,
    bottom: 0
  })
}

items.forEach(function(item, index) {
  item.addEventListener('click', el => {
    if (navigated === false) {
      indicator
        .css({
          width: `${selectedLi.offsetWidth}px`,
          left: `${selectedLi.offsetLeft}px`,
          bottom: 0
        })
        .detach()
        .appendTo('.navbar ul') //.appendTo(el.target)
    }
    navigated = true
    handleIndicator(item)
  })
})
