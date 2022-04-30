import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import Banner from './banner'
import { getAgent,sendCommand } from '../actions/c2actions'
import CommandBlock from './command';

const Agent = (props) => {
  const [result, setResult] = useState([]);
  const [loading, setLoading] = useState(true);
  let { id } = useParams();
  let textInput = React.createRef();

  const init = async () => {
    var res = await getAgent(id)
    console.log(res.payload)
    if (res.payload) {
      setResult(res.payload)
      setLoading(false)
    }
  }

  const send = async () => {
    var command = {
      command: textInput.current.value
    }
    var res = await sendCommand(id, command)
    console.log(res.payload)
  }

  useEffect(() => {
  //  init()
  setInterval(init(), 1000)
  },[]);

  if (loading) {
    return (
      <font size={2} color="#FFFFFF">
        Loading.....
      </font>
    );
  }
  return (
    <td colSpan={1} width={1} valign="top" height={1}>
      {/*Start Center Area*/}
      <center>
        <table
          width={1}
          height={1}
          cellSpacing={0}
          cellPadding={0}
          border={0}
        >
          {/*Start Title Row*/}
          <tbody>
            <Banner />
            {/* End title Row */}
            {/*Start Blurb Row*/}
            <tr>
              <td>
                {/*Start Blurb */}
                <font
                  size={1}
                  face="lucida console"
                  color="#FFFFFF"
                >
                  CALLBACKS
                  <br />
                </font>
                {/* End Blurb*/}
              </td>
            </tr>
            {/*End Blurb Row*/}
            <CommandBlock
              list={result}
            />
            {/*Start News 1 Banner */}
            <pre>
                    <font face="lucida console">
                        <font size={2} color="#ff3333">
                            {" "}
                            ______________________________________________________{" "}
                        </font>
                        {"\n"}
                        {/* <font size={2} color="#ff3333">
                            |
                        </font> */}
                        <font size={2} color="#888888">
                            {" "}
                            +{" "}
                        </font>
                        <font size={2} color="#FFFFFF">
                            {/*Start News 1 Title */}
                            {/* {this.props.command} */}
                            <input ref={textInput} type="text" /><button onClick={send}>
          Send
        </button>
                            {/*End News 1 Title */}
                        </font>
                        <font size={2} color="#888888">
                            {" "}
                            +{"     "}
                        </font>
                        <font size={2} color="#888888">
                            {/* <font size={2} color="#ff3333">
                                |
                            </font> */}
                            {"\n"}
                            <font size={2} color="#ff3333">
                                {" "}
                                ------------------------------------------------------{" "}
                            </font>
                        </font>
                    </font>
                </pre>
                {/*End News 1 Banner */}
          </tbody>
        </table>
      </center>
    </td>
  )

};
export default Agent