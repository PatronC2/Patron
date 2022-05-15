import React, { useState } from 'react';
import Banner from './banner'
import { genPayload } from '../actions/c2actions'

const CreatePayload = () => {
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [serverip, setServerIp] = useState('');
  const [serverport, setServerPort] = useState('');
  const [callbackfrequency, setCallbackfrequency] = useState('');
  const [callbackjitter, setCallbackjitter] = useState('');
  const [type, setType] = useState('');

  const send = async () => {
    var command = {
      name: name,
      description: description,
      serverip: serverip,
      serverport: serverport,
      callbackfrequency: callbackfrequency,
      callbackjitter: callbackjitter,
      type: type
    }
    console.log(command)
    var res = await genPayload(command)
    console.log(res.payload)
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
                  Generate Payload
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
                        Name
                            +{" "}
                        </font>
                        <font size={2} color="#FFFFFF">
                            {/*Start News 1 Title */}
                            {/* {this.props.command} */}
                            <input onChange={e => setName(e.target.value)} type="text" />
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
                        Description
                            +{" "}
                        </font>
                        <font size={2} color="#FFFFFF">
                            {/*Start News 1 Title */}
                            {/* {this.props.command} */}
                            <input onChange={e => setDescription(e.target.value)} type="text" />
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
                        Server IP
                            +{" "}
                        </font>
                        <font size={2} color="#FFFFFF">
                            {/*Start News 1 Title */}
                            {/* {this.props.command} */}
                            <input onChange={e => setServerIp(e.target.value)} type="text" />
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
                        Server Port
                            +{" "}
                        </font>
                        <font size={2} color="#FFFFFF">
                            {/*Start News 1 Title */}
                            {/* {this.props.command} */}
                            <input onChange={e => setServerPort(e.target.value)} type="text" />
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
                            <input onChange={e => setCallbackfrequency(e.target.value)} type="text" />
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
                            <input onChange={e => setCallbackjitter(e.target.value)} type="text" />
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
                        Payload Type
                            +{" "}
                        </font>
                        <font size={2} color="#FFFFFF">
                            {/*Start News 1 Title */}
                            {/* {this.props.command} */}
                            <select onChange={e => setType(e.target.value)}>
                              <option value="original">Shell</option>
                              <option value="wkeys">Shell + keylogger</option>
                            </select>
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
                              Generate Payload
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
export default CreatePayload