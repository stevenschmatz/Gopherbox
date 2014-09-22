==========
Initialization
==========
	1. Client sends server: username
	2. Client sends server: password
	3. Authenticate on server, WITH the RSA PUBLIC KEY to identify client
	4. If good, server sends client user’s activation key
	5. Server then sends client the Last_accessed time of the client
	6. Client decrypts secret key with activation key, stores it in a variable
	7. Client sends server files in FileMap with modification time GREATER THAN the Last_accessed time.
	8. Session opens

==========
Session
==========
	- Server, every 5 seconds, queries for files in UserFiles that have a modification time GREATER THAN client’s Last_accessed time, and sends those files. (This keeps files synced across multiple systems).
	- Client, every 5 seconds, sends files as instructed by the “watch” module.

==========
Method:
==========
-----BEGIN SEND DATA-----
-----BEGIN METADATA-----

-----BEGIN PATH-----
[path goes here]
-----END PATH——

-----BEGIN SIZE——
[size in bytes goes here]
——END SIZE-----

-----BEGIN IS_DIR——
[1 or 0 goes here]
-----END IS_DIR-----

——BEGIN MODIFICATION_TIME——
[modification time goes here]
——END MODIFICATION_TIME——
-----END METADATA-----

-----BEGIN CONTENTS-----
[contents go here]
-----END CONTENTS-----

-----END SEND DATA-----

==========
Termination
==========
	1.	Update the client’s Last_accessed time
	2.	If there was an incomplete data transfer on server or client, delete file


TODO: Implement better init system:
	1.	Server sends the client the list of files for that user.
	2.	Client has two operations going on:
		1.  Walk through all files on user, and if the file.path is NOT in the server list, add to list of files to send. If it IS in the server list, delete it from the server list.
		2.	The remaining files are the ones which must be sent to the client.
	3.	Client sends the server the list of files to send

TODO: Implement better file-uploading queue
	1.	Server saves file to EC2 temporarily, just to preserve file contents
	2.	Go func() { [upload to server and delete file on EC2] }()

TODO: Fix timeout problem

TODO: File synchronization across multiple machines





BUGS THAT SUCKED

- Fixing a data race earlier.
- Errors can produce mysterious effects in places you might not expect: Once the filepath communication (to sync clients) was implemented, all of a sudden the file sending from client to server was not working at all. Mysteriously, there were no errors produced. I implemented more rigorous error checking and found that the location of the error was not dealing with sending a file at all, but a necessary step before – unpadding the id_aes private key needed to encrypt the file contents. I then located the error specifically within the unpadding function, and it was borne from using a key of incorrect length. I thought this must be impossible, because I now had one client working with a key, and I used the same key with the other client - still not working. The only reasonable explanation was that the client was using a different key than the filepath given. I logged the working directory and filepath, and voila – it was using a different key than it should have been. Ultimately, this was due to an error in entering the directory of the source files. I now check for errors in the user input very carefully...