'use strict'

function Application() {
	this._socket = new WebSocket("ws://localhost:8080/ws")

	/**
	 * Object containing references to important dom elements.
	 * @type {Object}
	 */
	this._dom = {
		/**
		 * Element which will contain all messages.
		 * @type {HTML Element}
		 */
		conversation: document.getElementById("conversation"),

		/**
		 * Input element to be used to send new messages.
		 * @type {HTML Element}
		 */
		textbox: document.getElementById("textbox"),

		/**
		 * Element to display information about the number of online users.
		 * @type {HTML Element}
		 */
		count: document.getElementById("count")
	}

	this._init()
}

/**
 * Initializer.
 */
Application.prototype._init = function () {
	this._setupSocket()
		._setupTextbox()
}

/**
 * Connect with the websocket server and setup listeners.
 * @return {Application} The instance.
 */
Application.prototype._setupSocket = function () {
	var self = this
	this._socket.onmessage = function (resp) {
		var msg = JSON.parse(resp.data)

		switch (msg.type) {
			case "message": {
				self._onMessage(msg.data)
				break
			}
			case "login": {
				self._onLogin(msg.data)
				break
			}
			case "logout": {
				self._onLogout(msg.data)
				break
			}
		}
	}
	return this
}

/**
 * Add a listener on the Textbox element which when the user clicks on the Enter
 * key the text within the input field is sent to the websocket server.
 * @return {Application} The instance.
 */
Application.prototype._setupTextbox = function () {
	var self = this
	this._dom.textbox.onkeydown = function (ev) {
		if (ev.keyCode == 13 && this.value !== "") {
			self._socket.send(this.value)
			this.value = ""
		}
	}
	return this
}

/**
 * Method used to create a comment box.
 * @return {HTML Element} The comment box element.
 */
Application.prototype._createCommentBox = function () {
	var elem = document.createElement("div")
	elem.className = "message"
	return elem
}

/**
 * Method triggered when a new message is recieved from the server.
 * @param  {String} msg The message to be displayed.
 */
Application.prototype._onMessage = function (msg) {
	var elem = this._createCommentBox()
	elem.innerHTML = msg
	this._dom.conversation.appendChild(elem)
}

/**
 * Method triggered when the message recieved is of type 'login' which means a
 * new user has joined to conversation.
 * @param  {Number} Object.user 	The id of the user.
 * @param  {Number} Object.count 	The number of online users.
 */
Application.prototype._onLogin = function (msg) {
	var elem = this._createCommentBox()
	elem.className = "message login"
	elem.innerHTML = "User " + msg.user + " Logged in"
	this._dom.conversation.appendChild(elem)
	this._dom.count.innerHTML = msg.count
}

/**
 * Method triggered when the message recieved is of type 'logout' which means a
 * new user has exited to conversation.
 * @param  {Number} Object.user 	The id of the user.
 * @param  {Number} Object.count 	The number of online users.
 */
Application.prototype._onLogout = function (msg) {
	var elem = this._createCommentBox()
	elem.className = "message logout"
	elem.innerHTML = "User " + msg.user + " Logged out"
	this._dom.conversation.appendChild(elem)
	this._dom.count.innerHTML = msg.count
}

;(function () {
	var app = new Application()
}())