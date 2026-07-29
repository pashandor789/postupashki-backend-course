use std::net::{TcpListener, TcpStream, SocketAddr, Shutdown};
use std::io::{Result, BufWriter, Stdout, Write, stdout};

fn reply(conn: &TcpStream, out: &mut BufWriter<Stdout>) -> Result<()> {
    let reply_msg: &str = "OK\n";
    let mut stream: BufWriter<&TcpStream> = BufWriter::new(conn);
    stream.write_all(reply_msg.as_bytes())?;
    stream.flush()?;

    writeln!(out, "written \'{}\' to socket", &reply_msg)?;
    out.flush()?;
    
    return conn.shutdown(Shutdown::Write);
}

fn main() -> Result<()> {
    let mut out: BufWriter<Stdout> = BufWriter::new(stdout()); 

    let listener: TcpListener = match TcpListener::bind(SocketAddr::from(([127, 0, 0, 1], 8080)))
    {
        Ok(listener) => listener,
        Err(err) => {
            writeln!(out, "{}", err)?;
            out.flush()?;
            return Err(err);
        }
    };

    for stream  in listener.incoming() {
        match stream {
            Ok(guest) => reply(&guest, &mut out)?,
            Err(err) =>  {
                writeln!(out, "{}", err)?;
                out.flush()?;
                return Err(err);
            }
        }
    }

    return Ok(());
}
