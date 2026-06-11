# symmetrical-void

# About WebRTC in 100 seconds by Fireship
https://www.youtube.com/watch?v=WmR9IMUD_CY

Computer 1 sends and offer to Computer 2 asking to connect. 
The offer generates session description, defined by the session description protocol, describing the peer to peer connection.
The session description is saved in a signaling server (the websocket server in this example), which Computer 2 reads to generate an answer. 
The answer from Computer 2 is sent back to Computer 1 through the signaling server, confirming a connection can be established. 
Through the standard Interactive Connectivity Establishement, availible ports and ip addresses that can be used for the connection are exchanged as ICE candidates. 
WebRTC, through a STUN server, determines which ICE candidate is best for the connection.  



# WebRTC Go Package
https://github.com/pion/webrtc/
PAKE library for generating a strong secret between parties over an insecure channel


Here are some ideas to get your creative juices flowing:

Send a video file to multiple browser in real time for perfectly synchronized movie watching.
Send a webcam on an embedded device to your browser with no additional server required!
Securely send data between two servers, without using pub/sub.
Record your webcam and do special effects server side.
Build a conferencing application that processes audio/video and make decisions off of it.
Remotely control a robots and stream its cameras in realtime.

I'm thinking a collaborative Markdown doc.

# PAKE authentication Go Package
https://github.com/PeterJohnBishop/pake
PAKE library for generating a strong secret between parties over an insecure channel

