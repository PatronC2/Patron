import React, { Component } from 'react'
import { Link } from 'react-router-dom';
import Banner from './banner'
import PayloadStruct from './payloadstruct'
import {getPayloads} from '../actions/c2actions'


class Payloads extends Component {

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
    var res = await getPayloads()
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
            No Payload... <Link to={`/createpayload`}> Create Payload </Link>
          </font>
        );
      }
        return (
            <td colSpan={1} width={1} valign="top" height={1}>
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
                                  PAYLOADS <Link to={`/createpayload`}> Create Payload </Link>
                                  <br />
                                </font>
                                {/* End Blurb*/}
                              </td>
                            </tr>
                            {/*End Blurb Row*/}
                            <PayloadStruct
                              list={this.state.result}
                              />
                          </tbody>
                        </table>
                      </center>
                    </td>
        )
    }
}

export default Payloads