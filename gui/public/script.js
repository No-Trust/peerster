'use strict';

/*
    Constants
*/

const SERVER_ADDRESS = window.location.hostname;
const SERVER_PORT    = window.location.port;

const RUMOR_CHAT = 0;

const CHATS_TAB = document.getElementById('chats-tab');
const PEERS_TAB = document.getElementById('peers-tab');

const LEFT_PANE_LIST   = document.getElementById('left-pane-list');
const LEFT_PANE_INPUT  = document.getElementById('left-pane-input');
const LEFT_PANE_BUTTON = document.getElementById('left-pane-button');

const MESSAGE_LIST          = document.getElementById('message-list');
const MESSAGE_INPUT         = document.getElementById('message-input');
const FILE_INPUT            = document.getElementById('file-input');
const MESSAGE_ATTACH_BUTTON = document.getElementById('message-attach-button');
const MESSAGE_SEND_BUTTON   = document.getElementById('message-send-button');

const LEFT_PANE_INPUT_PLACEHOLDERS = Object.freeze({
    ENTER_NAME : 'Enter a name',
    ENTER_PEER : 'Enter a peer'
});

const SEND_MODES = Object.freeze({
    TEXT  : 0,
    FILES : 1
});

/*
    Variables
*/

// Own name
let myNames = [];

// Known peers
let knownPeers = [];

// Known origins
let chats = [
    'Rumor chat-room'
];

// Rumors and Messages
let rumors = {};
let messages = {};

let rumorReadIndexes = {};
let messageReadIndexes = {};

// UI
let activeChat;
let sendMode = SEND_MODES.TEXT;

/*
    Functions
*/

function removeChildren(element) {
    // Keep removing the first child of the
    // element until there are no more children
    while (element.hasChildNodes()) {
        element.removeChild(element.firstChild);
    }
}

function copyTextToClipboard(text) {

    const TEXTAREA = document.createElement('TEXTAREA');
    TEXTAREA.style.position = 'fixed';
    TEXTAREA.style.top = 0;
    TEXTAREA.style.left = 0;
    TEXTAREA.style.width = '2em';
    TEXTAREA.style.height = '2em';
    TEXTAREA.style.padding = 0;
    TEXTAREA.style.border = 'none';
    TEXTAREA.style.outline = 'none';
    TEXTAREA.style.boxShadow = 'none';
    TEXTAREA.style.background = 'transparent';
    TEXTAREA.value = text;

    document.body.appendChild(TEXTAREA);

    TEXTAREA.select();

    try {
        document.execCommand('copy');
    } catch (err) {
        console.error(err);
    }

    document.body.removeChild(TEXTAREA);

}

function chatsTabIsSelected() {
    return 'selected' in CHATS_TAB.dataset;
}

function peersTabIsSelected() {
    return 'selected' in PEERS_TAB.dataset;
}

function activateChat(chatName) {

    let chats = Array.from(LEFT_PANE_LIST.children);

    chats.forEach(chat => {
        delete chat.dataset.selected;
        if (chat.innerHTML === chatName) {
            chat.dataset.selected = '';
        }
    });

    if (chatName === RUMOR_CHAT) {
        LEFT_PANE_LIST.children[0].dataset.selected = '';
    }

    let newActiveChat = chatName;

    let differentChat = newActiveChat !== activeChat;

    if (differentChat) {
        removeChildren(MESSAGE_LIST);

        rumorReadIndexes = {};
        messageReadIndexes = {};
    }

    activeChat = newActiveChat;

    if (differentChat) {
        updateChat();
    }

}

function addMessage(origin, message) {
    const CONTAINER = document.createElement('DIV');
    CONTAINER.classList.add('message-containers');

    if (myNames.includes(origin)) {
        CONTAINER.classList.add('message-containers-from-me');
    }

    const MESSAGE = document.createElement('DIV');
    MESSAGE.classList.add('messages');

    const ORIGIN = document.createElement('P');
    ORIGIN.innerHTML = origin;
    ORIGIN.addEventListener('click', event => {
        if (chats.slice(1).includes(origin)) {
            activateChat(origin);
        }
    });

    const MESSAGE_TEXT = document.createElement('P');
    MESSAGE_TEXT.innerHTML = message;

    MESSAGE.appendChild(ORIGIN);
    MESSAGE.appendChild(MESSAGE_TEXT);

    CONTAINER.appendChild(MESSAGE);

    MESSAGE_LIST.appendChild(CONTAINER);

    MESSAGE_LIST.scrollTop = MESSAGE_LIST.scrollHeight;
}

function addPeer(peer) {
    const PEER = document.createElement('P');
    PEER.classList.add('cards', 'peer-cards');
    PEER.innerHTML = peer;

    PEER.addEventListener('click', event => {
        copyTextToClipboard(PEER.innerHTML);
        PEER.innerHTML = 'Copied!';
        setTimeout(() => PEER.innerHTML = peer, 1000);
    });

    LEFT_PANE_LIST.appendChild(PEER);
}

function addChat(origin) {
    const CHAT = document.createElement('P');
    CHAT.classList.add('cards', 'chat-cards');
    CHAT.innerHTML = origin;

    let chatName = (LEFT_PANE_LIST.children.length === 0) ? RUMOR_CHAT : origin;

    CHAT.addEventListener('click', event => activateChat(chatName));

    LEFT_PANE_LIST.appendChild(CHAT);
}

function updateChat() {

    if (activeChat === RUMOR_CHAT) {

        for (let origin in rumors) {
            if (!(origin in rumorReadIndexes)) {
                rumorReadIndexes[origin] = 0;
            }
            while (rumorReadIndexes[origin] < rumors[origin].length) {

                addMessage(origin, rumors[origin][rumorReadIndexes[origin]].Text);

                rumorReadIndexes[origin]++;

            }
        }

    } else {

        let origin = activeChat;

        if (!(origin in messages)) {
            // TODO: Add something to the GUI in this case
            console.log(`No messages with ${origin}`);
            return;
        }

        if (!(origin in messageReadIndexes)) {
            messageReadIndexes[origin] = 0;
        }

        while (messageReadIndexes[origin] < messages[origin].length) {

            let message = messages[origin][messageReadIndexes[origin]];
            addMessage(message.Origin, message.Text);

            messageReadIndexes[origin]++;

        }

    }

}

function switchToFileMode() {

    MESSAGE_INPUT.value = FILE_INPUT.files[0].name;
    MESSAGE_INPUT.disabled = true;

    sendMode = SEND_MODES.FILES;

}

function switchToTextMode() {

    MESSAGE_INPUT.value = '';
    MESSAGE_INPUT.disabled = false;

    sendMode = SEND_MODES.TEXT;

}

/*
    GETters
*/

function getRumors() {

    fetch(`http://${SERVER_ADDRESS}:${SERVER_PORT}/message`)
        .then(response => response.json())
        .then(data => {

            // myNames.push(data.Name);

            // rumors = data;
            data.forEach(message => {

                if (!(message.SenderName in rumors)) {
                    rumors.SenderName = [];
                }

                rumors.SenderName.push(message);

            });

            updateChat();

        }).catch(console.error);

}

function getMessages() {

    fetch(`http://${SERVER_ADDRESS}:${SERVER_PORT}/private-message`)
        .then(response => response.json())
        .then(data => {

            // myNames.push(data.Name);

            // messages = data;
            data.forEach(message => {

                if (!(message.Origin in messages)) {
                    messages.Origin = [];
                }

                messages.Origin.push(message);

            });

            updateChat();

        }).catch(console.error);

}

function getPeers() {

    fetch(`http://${SERVER_ADDRESS}:${SERVER_PORT}/node`)
        .then(response => response.json())
        .then(data => data.forEach(peer => {

            if (!knownPeers.includes(peer)) {
                knownPeers.push(peer);
                if (peersTabIsSelected()) {
                    addPeer(peer);
                }
            }

        })).catch(console.error);

}

function getOrigins() {

    fetch(`http://${SERVER_ADDRESS}:${SERVER_PORT}/reachable-node`)
        .then(response => response.json())
        .then(data => data.forEach(origin => {

            if (!chats.includes(origin)) {
                chats.push(origin);
                if (chatsTabIsSelected()) {
                    addChat(origin);
                }
            }

        })).catch(console.error);

}

/*
    POSTers
*/

function postMessage() {

    let message = MESSAGE_INPUT.value;
    let destination = (activeChat === RUMOR_CHAT) ? '' : activeChat;

    MESSAGE_INPUT.value = '';

    if (message !== '') {
        fetch(`http://${SERVER_ADDRESS}:${SERVER_PORT}/message`, {
            method : 'POST',
            body   : JSON.stringify({
                "message"     : message,
                "destination" : destination
            })
        }).catch(console.error);
    }

}

function postFile() {

    let filename = FILE_INPUT.files[0].name;

    if (filename !== '') {
        fetch(`http://${SERVER_ADDRESS}:${SERVER_PORT}/file`, {
            method : 'POST',
            body   : JSON.stringify({
                "filename" : filename
            })
        }).catch(console.error);
    }

}

function postPeer() {

    let peer = LEFT_PANE_INPUT.value;

    LEFT_PANE_INPUT.value = '';

    if (peer !== '') {
        fetch(`http://${SERVER_ADDRESS}:${SERVER_PORT}/node`, {
            method : 'POST',
            body   : JSON.stringify({
                "node" : peer
            })
        }).catch(console.error);
    }

}

function postDownload() {

    // TODO
    let hexhash, destination, filename;

    fetch(`http://${SERVER_ADDRESS}:${SERVER_PORT}/download`, {
        method : 'POST',
        body   : JSON.stringify({
            "hexhash"     : hexhash,
            "destination" : destination,
            "filename"    : filename
        })
    });

}

/*
    Listeners
*/

window.addEventListener('keypress', event => {
    if (event.keyCode === 13) {
        switch (document.activeElement) {
            case LEFT_PANE_INPUT:
                chatsTabIsSelected() ? postName() : postPeer();
                break;
            case MESSAGE_INPUT:
                postMessage();
                break;
        }
    }
});

CHATS_TAB.addEventListener('click', event => {
    delete PEERS_TAB.dataset.selected;
    CHATS_TAB.dataset.selected = '';

    removeChildren(LEFT_PANE_LIST);

    chats.forEach(origin => {
        addChat(origin);
    });

    activateChat(activeChat);

    LEFT_PANE_INPUT.placeholder = LEFT_PANE_INPUT_PLACEHOLDERS.ENTER_NAME;

    LEFT_PANE_BUTTON.classList.remove('mdi-account-plus');
    LEFT_PANE_BUTTON.classList.add('mdi-account-edit');
});

PEERS_TAB.addEventListener('click', event => {
    delete CHATS_TAB.dataset.selected;
    PEERS_TAB.dataset.selected = '';

    removeChildren(LEFT_PANE_LIST);

    knownPeers.forEach(peer => {
        addPeer(peer);
    });

    LEFT_PANE_INPUT.placeholder = LEFT_PANE_INPUT_PLACEHOLDERS.ENTER_PEER;

    LEFT_PANE_BUTTON.classList.remove('mdi-account-edit');
    LEFT_PANE_BUTTON.classList.add('mdi-account-plus');
});

LEFT_PANE_BUTTON.addEventListener('click', event =>
    chatsTabIsSelected() ? postName() : postPeer());

MESSAGE_ATTACH_BUTTON.addEventListener('click', event => FILE_INPUT.click());

FILE_INPUT.addEventListener('change', event => {

    if (FILE_INPUT.files.length > 0) {
        switchToFileMode();
    } else {
        switchToTextMode();
    }

});

MESSAGE_SEND_BUTTON.addEventListener('click', event => {

    if (sendMode === SEND_MODES.TEXT) {
        postMessage();
    } else {
        postFile();
        switchToTextMode();
    }

});

/*
    Execution
*/

chats.forEach(origin => {
    addChat(origin);
});

activateChat(RUMOR_CHAT);

setInterval(() => {
    getRumors();
    getMessages();
    getPeers();
    getOrigins();
}, 1000);
