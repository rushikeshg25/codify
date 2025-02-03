import { io, Socket } from "socket.io-client";
import { useEffect, useState } from "react";

const useSocket = (url: string) => {
  const [socket, setSocket] = useState<Socket | null>(null);

  useEffect(() => {
    if (!url) return;

    const newSocket = io(url);
    setSocket(newSocket);

    return () => {
      newSocket.disconnect();
    };
  }, [url]);
  return socket;
};

export default useSocket;
