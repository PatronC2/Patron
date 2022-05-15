import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import Banner from './banner'
import { getKeylog } from '../actions/c2actions'

const Keylog = (props) => {
  const [result, setResult] = useState([]);
  const [loading, setLoading] = useState(true);
  let { id } = useParams();
  let textInput = React.createRef();

   const getKey = async () => {
    var res = await getKeylog(id)
    console.log(res.payload)
    if (res.payload) {
      setResult(res.payload)
      setLoading(false)
    }
  }

  useEffect(() => {
  //  init()
  setInterval(getKey(), 1000)
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
                  KEYLOG
                  <br />
                </font>
                {/* End Blurb*/}
              </td>
            </tr>
            {/*End Blurb Row*/}
            {/*Start News 1 Banner */}
            <pre>
                    <font face="lucida console">
                        <font size={2} color="#00FFFF">
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
                            <textarea value={result[0].Keys ? result[0].Keys  : "No Keys..." } rows={30}/>
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
                            <font size={2} color="#00FFFF">
                                {" "}
                                ------------------------------------------------------{" "}
                            </font>
                        </font>
                    </font>
                </pre>
                {/*End News 1 Banner */}
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
                           <button onClick={getKey}>
                            Refresh
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
export default Keylog