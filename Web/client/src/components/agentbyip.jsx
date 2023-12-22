import React, { useEffect } from 'react'
import Banner from './banner'
import { getIps } from '../actions/c2actions'
import AgentByIpRow from './agentbyiprow';


const AgentByIp = () => {

  const [state, setState] = React.useState({
    result: [],
    loading: true,
  });

  useEffect(() => {
    const init = async () => {
      var res = await getIps();
      console.log(res.payload);
      if (res.payload) {
        setState({
          ...state,
          result: res.payload,
          loading: false,
        });
        console.log(state);
      }
    };

    init();

    // If you want to periodically refresh the data, you can use setInterval
    // const intervalId = setInterval(init, 1000);

    // Cleanup the interval when the component unmounts
    // return () => clearInterval(intervalId);
  }, []); // Only run this effect once when the component mounts
        if (state.loading) {
          return (
            <font size={2} color="#FFFFFF">
              No Agent...
            </font>
          );
        }
        return (
            <td className="stuck" colSpan={1} width={1} valign="top" height={1}>
                      {/*Start Center Area*/}
                      <center>
                        {console.log(state)}
                        <table
                          width={1}
                          height={1}
                          cellSpacing={0}
                          cellPadding={0}
                          border={0}
                        >
                          {/*Start Title Row*/}
                          <tbody>
                              <Banner/>
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
                                  IPs
                                  <br />
                                </font>
                                {/* End Blurb*/}
                              </td>
                            </tr>
                            {/*End Blurb Row*/}
                             {state.result?.map((item) => (
                                <AgentByIpRow
                                  key={item.agentip}
                                  agentip={item.agentip}
                                  // Pass other props as needed
                                />
                              ))}
                          </tbody>
                        </table>
                      </center>
                    </td>
          );
};

export default AgentByIp