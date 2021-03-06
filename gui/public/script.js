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

const MESSAGE_LIST            = document.getElementById('message-list');
const MESSAGE_INPUT           = document.getElementById('message-input');
const FILE_INPUT              = document.getElementById('file-input');
const MESSAGE_ATTACH_BUTTON   = document.getElementById('message-attach-button');
const MESSAGE_SEND_BUTTON     = document.getElementById('message-send-button');
const MESSAGE_DOWNLOAD_BUTTON = document.getElementById('message-download-button');

const DOWNLOAD_DIALOG_CONTAINER      = document.getElementById('download-dialog-container');
const DOWNLOAD_DIALOG_FILENAME_INPUT = document.getElementById('download-dialog-filename-input');
const DOWNLOAD_DIALOG_HEXHASH_INPUT  = document.getElementById('download-dialog-hexhash-input');
const DOWNLOAD_DIALOG_DEST_INPUT     = document.getElementById('download-dialog-dest-input');
const DOWNLOAD_DIALOG_ORIGIN_INPUT   = document.getElementById('download-dialog-origin-input');
const DOWNLOAD_DIALOG_BUTTON         = document.getElementById('download-dialog-button');

const SEND_MODES = Object.freeze({
    TEXT  : 0,
    FILES : 1
});

const REP_PREC = 3;

/*
    Variables
*/

// Own name
let myNames = [];

// Known peers
let knownPeers = [];

// Known origins
let chats = [
    'Chat-room'
];

// Rumors and Messages
let rumors = {};
let messages = {};

let rumorReadIndexes = {};
let messageReadIndexes = {};

// Reputations
let reputations = {
    SigReps     : {},
    ContribReps : {}
};

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

function clamp(value, min, max) {
    return (value <= min) ? min : (value >= max) ? max : value;
}

function remapValueInDomain(value, inMin, inMax, outMin, outMax) {
    return ((clamp(value, inMin, inMax) - inMin) / (inMax - inMin)) * (outMax - outMin) + outMin;
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
        if (chat.children[0].innerHTML === chatName) {
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

    function copy() {
        clearTimeout(timeout);
        copyTextToClipboard(PEER.innerHTML);
        PEER.innerHTML = PEER.innerHTML.endsWith('Copied!') ? 'Copied the copy!' :
            ((PEER.innerHTML.length > 70) || (PEER.innerHTML === 'COPYCEPTION!')) ? 'COPYCEPTION!' :
            !PEER.innerHTML.endsWith('copy!') ? 'Copied!' :
            `${PEER.innerHTML.slice(0, PEER.innerHTML.length - 1)} of the copy!`;
        timeout = setTimeout(() => PEER.innerHTML = peer, 1000);
    }

    let timeout;

    const CARD = document.createElement('DIV');
    CARD.classList.add('cards', 'peer-cards');

    const PEER = document.createElement('P');
    PEER.classList.add('titles');
    PEER.innerHTML = peer;

    const REP = document.createElement('P');
    REP.classList.add('reputations');
    REP.innerHTML = '-';

    CARD.addEventListener('click', copy);

    CARD.appendChild(PEER);
    CARD.appendChild(REP);
    LEFT_PANE_LIST.appendChild(CARD);

}

function addChat(origin) {

    const CARD = document.createElement('DIV');
    CARD.classList.add('cards', 'peer-cards');

    const CHAT = document.createElement('P');
    CHAT.classList.add('titles');
    CHAT.innerHTML = origin;

    const REP = document.createElement('P');
    REP.classList.add('reputations');
    REP.innerHTML = '-';

    let chatName = (LEFT_PANE_LIST.children.length === 0) ? RUMOR_CHAT : origin;

    CARD.addEventListener('click', event => activateChat(chatName));

    CARD.appendChild(CHAT);
    CARD.appendChild(REP);
    LEFT_PANE_LIST.appendChild(CARD);
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

function updateReputations() {

    let reps = chatsTabIsSelected() ? reputations.SigReps : reputations.ContribReps;

    if (reps === null) { return; }

    Array.from(LEFT_PANE_LIST.children).forEach(peer => {

        const ADDRESS   = peer.children[0];
        const REP_FIELD = peer.children[1];

        if (ADDRESS.innerHTML in reps) {

            const REP = reps[ADDRESS.innerHTML];

            const RED   = Math.round(remapValueInDomain(- REP, -1, -0.5, 0, 255));
            const GREEN = Math.round(remapValueInDomain(  REP,  0,  0.5, 0, 255));
            REP_FIELD.style.setProperty('color', `rgb(${RED}, ${GREEN}, 0)`);
            REP_FIELD.innerHTML = parseFloat(
                `${Math.trunc(REP * (10 ** REP_PREC)) * (10 ** - REP_PREC)}`.slice(0, 2 + REP_PREC));

        }

    });

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

function showDownloadDialog() {

    DOWNLOAD_DIALOG_CONTAINER.style.setProperty('display', 'inline');

}

function hideDownloadDialog() {

    DOWNLOAD_DIALOG_CONTAINER.style.setProperty('display', 'none');

}

/*
    GETters
*/

function getRumors() {

    fetch(`http://${SERVER_ADDRESS}:${SERVER_PORT}/message`)
        .then(response => response.json())
        .then(data => {

            // myNames.push(data.Name);

            if (data !== null) {
                rumors = [];
    
                data.forEach(message => {
    
                    if (!(message.SenderName in rumors)) {
                        rumors[message.SenderName] = [];
                    }
    
                    rumors[message.SenderName].push(message);
    
                });
    
                updateChat();
            }

        }).catch(console.error);

}

function getMessages() {

    fetch(`http://${SERVER_ADDRESS}:${SERVER_PORT}/private-message`)
        .then(response => response.json())
        .then(data => {

            // myNames.push(data.Name);

            if (data !== null) {
                messages = [];

                data.forEach(message => {
    
                    if (!(message.Origin in messages)) {
                        messages[message.Origin] = [];
                    }
    
                    messages[message.Origin].push(message);
    
                });
    
                updateChat();
            }

        }).catch(console.error);

}

function getPeers() {

    fetch(`http://${SERVER_ADDRESS}:${SERVER_PORT}/node`)
        .then(response => response.json())
        .then(data => data.forEach(peer => {

            let peerString = `${peer.Address.IP}:${peer.Address.Port}`;
            if (!knownPeers.includes(peerString)) {
                knownPeers.push(peerString);
                if (peersTabIsSelected()) {
                    addPeer(peerString);
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

function getReputations() {

    fetch(`http://${SERVER_ADDRESS}:${SERVER_PORT}/reputations`)
        .then(response => response.json())
        .then(data => {

            reputations = data;

            updateReputations();

        }).catch(console.error);

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

    if (peer !== '') {

        fetch(`http://${SERVER_ADDRESS}:${SERVER_PORT}/node`, {
            method : 'POST',
            body   : JSON.stringify({
                "node" : peer
            })
        }).catch(console.error);

        LEFT_PANE_INPUT.value = '';

    }

}

function postDownload() {

    let filename    = DOWNLOAD_DIALOG_FILENAME_INPUT.value;
    let hexhash     = DOWNLOAD_DIALOG_HEXHASH_INPUT.value;
    let destination = DOWNLOAD_DIALOG_DEST_INPUT.value;
    let origin      = DOWNLOAD_DIALOG_ORIGIN_INPUT.value;

    if ((filename !== '') && (hexhash !== '') && (destination !== '') && (origin !== '')) {

        fetch(`http://${SERVER_ADDRESS}:${SERVER_PORT}/download`, {
            method : 'POST',
            body   : JSON.stringify({
                "filename"    : filename,
                "hexhash"     : hexhash,
                "destination" : destination,
                "origin"      : origin
            })
        });

        DOWNLOAD_DIALOG_FILENAME_INPUT.value = '';
        DOWNLOAD_DIALOG_HEXHASH_INPUT.value  = '';
        DOWNLOAD_DIALOG_DEST_INPUT.value     = '';
        DOWNLOAD_DIALOG_ORIGIN_INPUT.value   = '';

        hideDownloadDialog();

    }

}

/*
    Listeners
*/

window.addEventListener('keypress', event => {

    if (event.keyCode === 13) {

        switch (document.activeElement) {

            case LEFT_PANE_INPUT:
                postPeer();
                break;

            case MESSAGE_INPUT:
                postMessage();
                break;

            case DOWNLOAD_DIALOG_FILENAME_INPUT:
            case DOWNLOAD_DIALOG_HEXHASH_INPUT:
            case DOWNLOAD_DIALOG_DEST_INPUT:
            case DOWNLOAD_DIALOG_ORIGIN_INPUT:
                postDownload();
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

});

PEERS_TAB.addEventListener('click', event => {

    delete CHATS_TAB.dataset.selected;
    PEERS_TAB.dataset.selected = '';

    removeChildren(LEFT_PANE_LIST);

    knownPeers.forEach(peer => {
        addPeer(peer);
    });

});

LEFT_PANE_BUTTON.addEventListener('click', postPeer);

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

MESSAGE_DOWNLOAD_BUTTON.addEventListener('click', showDownloadDialog);

Array.from(document.getElementsByClassName('dialog-dark-backgrounds'))
    .forEach(background => background.addEventListener('click', hideDownloadDialog));

DOWNLOAD_DIALOG_BUTTON.addEventListener('click', postDownload);

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
    getReputations();
}, 1000);
