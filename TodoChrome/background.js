chrome.browserAction.onClicked.addListener(function(tab) {
    $.post( "http://127.0.0.1:39095/insert", { title: tab.title, url: tab.url }, function( data ) {
//        chrome.notifications.create(NotificationOptions options, function callback)
      chrome.notifications.create('taskcreator', {
          type: 'basic',
          iconUrl: 'icon_128.png',
          title: 'Task created',
          message: tab.title }, 
        function(id) {
          timer = setTimeout(function(){
            chrome.notifications.clear(id);

          }, 3000);
        }
      );
// https://developer.chrome.com/apps/notifications#event-onClicked
      // chrome.notifications.onClicked.addListener(function callback)
//      chrome.notifications.onClicked.addListener(function(notificationId, byUser) {
//        chrome.tabs.create({url: "http://www.google.com"});
//      });
    } );

});
