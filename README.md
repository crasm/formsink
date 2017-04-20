formsink
========

*A standalone server for collecting POSTed forms from static sites into a Maildir.*

**Formsink is pre-alpha and has not been used in production! Don't use
this unless you want to help out with development.**

Forms on static sites is a pretty common problem, so there are a ton of
solutions already out there, such as [Formspree][], [Fireform][], and an
unofficial use of [Google Forms][].

[Formspree]: https://formspree.io/
[Fireform]: http://fireform.org/
[Google Forms]: https://github.com/toperkin/staticFormEmails

While [Formspree is open source][], self-hosting it still requires
sending the emails in plaintext through a [third-party service][].

[Formspree is open source]: https://github.com/formspree/formspree
[third-party service]: https://github.com/formspree/formspree#running-your-own-copy-of-formspree

Instead of sending emails, formsink lets you host the mailbox. If you
configure TLS on your website, formsink, and your email server, the form
contents are never sent over the wire in plaintext.

### Notes
- Email is still stored unencrypted in the maildir.

How it works
------------

Formsink parses html files to identify forms and know which form values
to grab from POST requests.

For instance, the following html...

```
<!DOCTYPE html>
<title>Contact Form</title>
<form method='post' action='http://localhost:1234/contact' enctype='multipart/form-data'>
	<ol>
		<li><label>name   <input type='text' name='name'/>    </label></li>
		<li><label>email  <input type='email' name='email'/>  </label></li>
		<li><label>message<textarea name='message'></textarea></label></li>
		<li><label>picture<input type='file' name='picture'/> </label></li>
	</ol>
	<input type='submit' value='Submit'/>
</form>
```

...becomes (in golang)...

```
&Form{
	Name:   "contact",
	Fields: []string{"name", "email", "message"},
	Files:  []string{"picture"},
}
```

When a POST request is received, formsink uses the `&Form{}` to build an
email and deposits the result in a maildir and a redirect is sent as a
response.

An email server detects the new email in the maildir and pushes it to
connected email clients.

Recommended setup
-----------------

I recommend using [dovecot][] to serve the maildir. You can connect to
dovecot with pretty much any email client.

[dovecot]: http://dovecot.org/
