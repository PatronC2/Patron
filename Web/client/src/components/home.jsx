import React, { Component } from 'react'
import Banner from './banner'
import { getCallbacks, deleteAgent, killAgent } from '../actions/c2actions'
import Row from './row';


const Home = () => {

  const [state, setState] = React.useState({
    result: [],
    selectedIds: [],
    loading: true,
  });

  React.useEffect(() => {
    const init = async () => {
      var res = await getCallbacks();
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

  const handleCheckboxChange = (uuid) => {
    setState((prevState) => {
      const isSelected = prevState.selectedIds.includes(uuid);
      let updatedSelectedIds;

      if (isSelected) {
        updatedSelectedIds = prevState.selectedIds.filter((id) => id !== uuid);
      } else {
        updatedSelectedIds = [...prevState.selectedIds, uuid];
      }

      return {
        ...prevState,
        selectedIds: updatedSelectedIds,
      };
    });
  };

  const kill = async () => {
    const { selectedIds } = state;

    if (selectedIds.length === 0) {
      return; // Do nothing if the selected IDs array is empty
    }

    for (const id of selectedIds) {
      var res = await killAgent(id);
      console.log(res.payload);
      if (res.payload === 'Success') {
        console.log('redirect');
      } else {
        console.log(res.payload);
      }
    }
    location.reload(true);
  };

  const deleteAg = async () => {
    const { selectedIds } = state;

    if (selectedIds.length === 0) {
      return; // Do nothing if the selected IDs array is empty
    }

    for (const id of selectedIds) {
      var res = await deleteAgent(id)
      console.log(res.payload)
      if (res.payload === "Success"){
        console.log("Success "+id)
      }else {
        console.log(res.payload)
      }
    }
    location.reload(true);
  };

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
                                  CALLBACKS
                                  <br />
                                </font>
                                <button onClick={kill}>Kill Selected Agents</button>
                                <button onClick={deleteAg}>Delete Selected Agents</button>
                                {/* End Blurb*/}
                              </td>
                            </tr>
                            {/*End Blurb Row*/}
                             {state.result?.map((item) => (
                                <Row
                                  key={item.uuid}
                                  uuid={item.uuid}
                                  username={item.username}
                                  status={item.status}
                                  agentip={item.agentip}
                                  isSelected={state.selectedIds.includes(item.uuid)}
                                  onCheckboxChange={handleCheckboxChange}
                                  // Pass other props as needed
                                />
                              ))}
                          </tbody>
                        </table>
                      </center>
                    </td>
          );
};

export default Home