import { io } from "socket.io-client";

const socket = io("http://api.rushikesh.localhost");
export default socket;
