var ids = []
dest = ""

setInterval(function() {
  // Updates messages every second
  updateMessages()
}, 1 * 1000);

setInterval(function() {
  // Updates reachable nodes every second
  updatePrivateMessages()
}, 1 * 1000);

setInterval(function() {
  // Updates nodes every second
  updateNodes()
}, 1 * 1000);

setInterval(function() {
  // Updates reachable nodes every second
  updateReachableNodes()
}, 1 * 1000);


function newPM() {
  dest = $('#reachable-nodes-select').find(":selected").text();
  if (dest != "") {
    $('#pm-text-div').show()
  }
}

function sendPM() {
  var msgText = $('#user_private_msg').val()
  $.post("message", {
    dest: dest,
    msg: msgText
  })
  $('#pm-text-div').hide()
  out = "You set to " + dest + " the message : " + msgText
  $('#privatemessage-list').append("<li class='message'>" + out + "</li>")
}

function sendMessage() {
  var msgText = $('#user_msg').val()
  $.post("message", {
    msg: msgText
  })
}

function addNode() {
  var newNode = $('#new_node').val()
  $.post("node", {
    node: newNode
  })
}

function changeName() {
  var newName = $('#peer_name').val()
  $('#peer_name').attr("placeholder", newName)
  $('#current_name').text(newName)
  $.post("id", {
    name: newName
  })
}

function submitFile() {
  var input = document.getElementById("file").files[0].name

  // only send top level filename
  $.post("file", {
    filename: input
  })
}

function downloadFile() {
  var hexhash = $('#filehash').val()
  var from = $('#reachable-nodes-select2').find(":selected").text();
  var filename = $('#dl_filename').val()
  // send request
  $.post("download", {
    MetaHash: hexhash,
    Destination: from,
    FileName: filename
  })
}

function changeNameWith(newName) {
  $('#peer_name').attr("placeholder", newName)
  $('#current_name').text(newName)
  $.post("id", {
    name: newName
  })
}

function updateMessages() {
  $.getJSON("message", function(data) {
    if (data != null) {
      $('#message-list').text("")
      $.each(data, function(k, v) {
        var sender = "<span class='sender_name'>" + sanitize(v.SenderName) + "</span>"
        var text = "<span class='message_text'>" + sanitize(v.Text) + "</span>"
        var out = sender + "<br />" + text
        $('#message-list').append("<li class='message'>" + out + "</li>")
      });
    }
  });
}

function updatePrivateMessages() {
  $.getJSON("private-message", function(data) {
    if (data != null) {
      $('#privatemessage-list').text("")
      $.each(data, function(k, v) {
        var sender = "<span class='sender_name'>" + sanitize(v.Origin) + "</span>"
        var text = "<span class='message_text'>" + sanitize(v.Text) + "</span>"
        var out = sender + "<br />" + text
        $('#privatemessage-list').append("<li class='message'>" + out + "</li>")
      });
    }
  });
}

function updateNodes() {
  $.getJSON("node", function(data) {
    if (data != null) {
      $('#node-list').text("")
      $.each(data, function(k, v) {
        if (v != null) {
          for (var i = 0; i < v.length; i++) {
            var peer = v[i]
            var ip = peer.Address.IP
            var port = peer.Address.Port
            var id = peer.Identifier
            var out = "<span class='node_address'>" + ip + ":" + port + "</span>" + "<br />" + id
            $('#node-list').append("<li class='node'>" + out + "</li>")
          }
        }
      });
    }
  });
}

function updateReachableNodes() {
  $.getJSON("reachable-node", function(data) {
    if (data != null) {
      if (!arraysEqual(data, ids)) {
        console.log(data)
        ids = data
        $('#reachable-nodes-select').text("")
        $.each(ids, function(k, v) {
          if (v != null) {
            var id = sanitize(v)
            var out = '<option value="' + id + '">' + id + '</option>'
            $('#reachable-nodes-select').append(out)
          }
        });
        $('#reachable-nodes-select2').text("")
        $.each(ids, function(k, v) {
          if (v != null) {
            var id = sanitize(v)
            var out = '<option value="' + id + '">' + id + '</option>'
            $('#reachable-nodes-select2').append(out)
          }
        });
      }
    }
  });
}

function arraysEqual(a, b) {
  if (a === b) return true;
  if (a == null || b == null) return false;
  if (a.length != b.length) return false;

  a = a.sort()
  b = b.sort()

  for (var i = 0; i < a.length; ++i) {
    if (a[i] !== b[i]) return false;
  }
  return true;
}



function sanitize(text) {
  return $($.parseHTML(text)).text();
}
