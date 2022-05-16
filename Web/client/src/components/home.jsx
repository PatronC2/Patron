import React, { Component } from 'react'
import Banner from './banner'
import Callback from './callback'
import {getCallbacks} from '../actions/c2actions'


class Home extends Component {

  constructor(props) {
    super(props);
    this.state = {
      result: [],
      loading: true
    }
  }

  componentDidMount = async () => {
    this.init()
    // setInterval(this.init(), 1000)
  }

  init = async () => {
    var res = await getCallbacks()
    console.log(res.payload)
    if(res.payload){
      this.setState({
        result: res.payload,
        loading: false
      })
      console.log(this.state)
    }  
  }
  

    render(){
      if (this.state.loading) {
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
                        {console.log(this.state)}
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
                                {/* End Blurb*/}
                              </td>
                            </tr>
                            {/*End Blurb Row*/}
                            <Callback
                              list={this.state.result}
                              />
                          </tbody>
                        </table>
                      </center>
                    </td>
        )
    }
}

export default Home