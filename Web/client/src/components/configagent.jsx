import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import Banner from './banner'
import { sendConfig, getOneAgent } from '../actions/c2actions'

const ConfigAgent = () => {
    const [errormsg, setError] = useState('');
  const [callbackserver, setCallbackServer] = useState('');
  const [callbackfrequency, setCallbackfrequency] = useState('');
  const [callbackjitter, setCallbackjitter] = useState('');
  const [loading, setLoading] = useState(true);
  let { id } = useParams();

  const send = async () => {
    var command = {
      callbackserver: callbackserver,
      callbackfreq: callbackfrequency,
      callbackjitter: callbackjitter
    }
    var res = await sendConfig(id,command)
    console.log(res.payload)
    if (res.payload === "Success"){
        // history.push('/')
        setError('')
        console.log('redirect')
      }else {
        setError(res.payload)
      }
  }

  const init = async () => {
    var res = await getOneAgent(id)
    console.log(res.payload)
    if (res.payload) {
        setCallbackServer(res.payload[0].callbackto)
        setCallbackfrequency(res.payload[0].callbackfrequency)
        setCallbackjitter(res.payload[0].callbackjitter)
        setLoading(false)
    }
  }

  useEffect(() => {
    init()
    },[]);
    if (loading) {
    return (
      <font size={2} color="#FFFFFF">
        Loading.....
      </font>
    );
  }

  var error = () => (
    <center>
                <pre>
                  <font face="lucida console">
                    <font size={2} color="#ff3333">
                      ________________________________________
                    </font>
                    {"\n"}
                    <font size={2} color="#ff3333">
                      |
                    </font>
                    <font size={2}>{"    "}</font>
                    <font size={2} color="#ff3333">
                      {" "}
                      +{" "}
                    </font>
                    <font size={2} color="#ff3333">
                      {" Error: "}{errormsg}{"    "}
                    </font>
                    <font size={2} color="#ff3333">
                      {" "}
                      +{"     "}
                    </font>
                    <font size={2} color="#ff3333">
                      |
                    </font>
                    {"\n"}
                    <font size={2} color="#ff3333">
                      -----------------------------------------
                    </font>
                  </font>
                  {"\n"}
                </pre>
              </center>
  )
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
                { errormsg ? error() : null }
                <font
                  size={1}
                  face="lucida console"
                  color="#FFFFFF"
                >
                  Configure Agent
                  <br />
                </font>
                {/* End Blurb*/}
              </td>
            </tr>
            {/*End Blurb Row*/}
            {/*Start News 1 Banner */}
            <pre>
                    <font face="lucida console">

                        {/*   begin */}
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
                        Callback Server:port
                            +{" "}
                        </font>
                        <font size={2} color="#FFFFFF">
                            {/*Start News 1 Title */}
                            {/* {this.props.command} */}
                            <input onChange={e => setCallbackServer(e.target.value)} value={callbackserver} type="text" />
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
                        <br/>
                        {/*   end */}

                        {/*   begin */}
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
                        Callback Frequency
                            +{" "}
                        </font>
                        <font size={2} color="#FFFFFF">
                            {/*Start News 1 Title */}
                            {/* {this.props.command} */}
                            <input onChange={e => setCallbackfrequency(e.target.value)} value={callbackfrequency} type="text" />
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
                        <br/>
                        {/*   end */}

                        {/*   begin */}
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
                        Callback Jitter
                            +{" "}
                        </font>
                        <font size={2} color="#FFFFFF">
                            {/*Start News 1 Title */}
                            {/* {this.props.command} */}
                            <input onChange={e => setCallbackjitter(e.target.value)} value={callbackjitter} type="text" />
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
                        <br/>
                        {/*   end */}
                        <br/>
                        <button onClick={send}>
                              Update
                        </button>
                    </font>
                </pre>
                {/*End News 1 Banner */}
          </tbody>
        </table>
      </center>
    </td>
  )

};
export default ConfigAgent