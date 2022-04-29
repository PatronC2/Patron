import React, { Component } from 'react'

class Banner extends Component {

    render(){
        return (
            
            <tr>
            <td>
              {/*Begin Main Title, Use menu.php to make these easily */}
              <center>
                <pre>
                  <font face="lucida console">
                    <font size={2} color="#00FFFF">
                      ________________________________________
                    </font>
                    {"\n"}
                    <font size={2} color="#00FFFF">
                      |
                    </font>
                    <font size={2}>{"    "}</font>
                    <font size={2} color="#FFC200">
                      {" "}
                      +{" "}
                    </font>
                    <font size={2} color="#FFFFFF">
                      {"   "}P a t r o n{"      "}C 2{"    "}
                    </font>
                    <font size={2} color="#FFC200">
                      {" "}
                      +{"     "}
                    </font>
                    <font size={2} color="#00FFFF">
                      |
                    </font>
                    {"\n"}
                    <font size={2} color="#00FFFF">
                      -----------------------------------------
                    </font>
                  </font>
                  {"\n"}
                </pre>
              </center>
              {/*End Main Title*/}
            </td>
          </tr>
        )
    }
}

export default Banner