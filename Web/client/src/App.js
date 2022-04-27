import { BrowserRouter, BrowserRouter as Router, Switch, Route, Link, NavLink } from 'react-router-dom';
import Header from './components/header'
import createDOMPurify from 'dompurify'
import { JSDOM } from 'jsdom'

const window = (new JSDOM('')).window
const DOMPurify = createDOMPurify(window)

var rawHTML =
  `
<center>
<!-- ## BEGIN ALL ## -->
<table border="0" cellspacing="0" cellpadding="0">
<!-- ## BEGIN UPPER AREA## -->
<tr><td>
<center>
<table border="0" cellspacing="0" cellpadding="0">
<tr><td colspan="1"><pre><font color="#00FF00" face="lucida console" size="1">                                                     </font></td><td><pre><font color="#00FF00" face="lucida console" size="1">                                                                                 </font></td><td><pre><font color="#00FF00" face="lucida console" size="1">                                                      </font></td></tr>
<tr><td colspan="1"><!--Begin Left Leg, make sure <font>, <pre> and first line of ascii are on same line--><pre><font color="#00FF00" face="lucida console" size="1">,@hiiiSiiiiX#   ...   ,rir ,@@@ .AXSirsXs2&amp;2#@
  @3iiiiiiSA:  .     r2ir. 2@iX..@A#S;;A#@@@H@
  .@3SSSSSSA:  ....:S2ir:;;@r2;:2@S;iir@@@B@X@
   :@XSSSSS3@   ,;rSisr;r2rA :i;,2HBAh3M@@9@#M
    r@2SSSSSMS    ,sisriSX;: :5:r;B@@@#@95#@@;
     2@2SSSSX@     ;SSSSXSs.,:;;rsr@@@@@&amp;A@9@,
      &amp;M2SSiSH3     siiS3 i.,;;;s9B###@AMM@2@
       #M5SSi2@     ;ssih.r..r:A@@#M#@MHMM@X@
        #H2SSiAH    :rsiX2 r ;@@AAAHB#BHMM@A@
         H35222@    .;rsSA ;, @@AHMB#@##@@@M@
          hiri5H@    ,;siGs S,X@H&amp;M#@@@@HG2S@
           3r:;;B.   .;rs2&amp; ...@@AAM#@2S293A@
            2;::r2     ;i2H, X.2@&amp;&amp;H#@h3GAhA@
            .s.  ,      :5As r;.@@GHB@AXAA&amp;H@
             r&amp;sssX@     :Si,,s A@3AA@#SA&amp;AA@
             .@Gh93@.    .ss:.r.G@GAH@#iAAAA@
              @2SSS@,     ;r;.;.&amp;@HAH@#sAAAA@&amp;
              @G5S5@;     :r;,;.H@@GH@#sAA&amp;AH@
              @#X22#2     .rr,. A@@3A@#sGA&amp;AA@
              2@hX2H#      rs, &amp;&amp;M@2A##sGA&amp;A&amp;@
              :@Gh9A@      ss, MhA@3A##s3A&amp;A&amp;@r
               @Ghh&amp;@      ii, BA3@HG##s2A&amp;G&amp;#@
               @Ahhh@      iS, MHX@#X#Mi2AGG&amp;A@
               @B339@      SS, A#3@@2#Mi2A&amp;G&amp;h@
               X@2X3@      2i. XMA#@XM#S2&amp;&amp;GG9@S
               .@222@      2i. iBMA@9H#S29GhG9@@
                @iS5#,     5s  :M@X@AA#SX9&amp;&amp;GG#@
                @isSHr     Sr   #@5@#h@5X3&amp;&amp;hhH@
                M2sS&amp;S     rr.  #@S@@9@232&amp;&amp;h9M@;
                S&amp;iSh9     .,.. #@5@@3@X32A&amp;G3M@@
                ,#i59#      .,, H@2#@3@3X2&amp;&amp;&amp;3#M@
                 @SXh@      .,, i@AH@G#hX2G&amp;G3#B@
                 @23G@      .,, ,@@G@A#A25hA&amp;X@B@;
                 @G9A@      .,, .#@2@MBH22hA&amp;X@BM@
                 3@GH@       ,,  M@5@@AM259A&amp;X@HG@
                 ,@AB@       ... @@5@@&amp;@253AG&amp;@h3@
                  @A&amp;@;    .  .  @@2@@G#2S3A&amp;AM#X@;
                  @Ah@&amp;    :,:,  B@h#@&amp;@5SXAGA#BG#@
                  @B3#@    :,;r: &amp;@MG@M#5iX&amp;h#&amp;GA9@
                  @@3H@    ,,s3r.,&amp;@A@@BSiX&amp;hM2AAS@
                  ;@Xh@    ..2A2:.rM2@@BiiXG3#B@Ai@;
                   @2X@     .229, s@.@@MrS2GG#X@&amp;i@@
                   @X2@     ,2SG, :@:&amp;@B;52hAA2@HiM@
                   @X2@     :55h. .@rr@h;529Gh3@HsG@
                   B3X@     ;iS&amp;. ;@X:@X;223&amp;9h@#S3@.
                   rG2@:    rsSG, 5@#,@S;22XG3h@#XS@@
                    H5@r   .;ri9r &amp;@@.@sr222h&amp;h@@32@@
                    Hr9;    :ri2XrS@@.@rr5222Ah@#5S@@
                    @GM@    ,;riSi:@@,@rs225S@G@3ih@,
                    r@@@    :ri2XX;@XsM2;32sB@G@3i#@
                     @@@    :s2GHHX@&amp;s#H.h5S@@&amp;@25@@
                     ,@@3   ;2GBMMH@@iX.,SS@@@A@i&amp;@
                      @@@   iAAHHM#@Hi&amp;:;;A@@@B@;#@
                      ,@@B  .A#M#@&amp;&amp;s;3rr;@@@#M@r@s
                       @@@   ,#@@h,s:r3Si;AGMMB@i@
                       ,@@9  , @.  Srr2S2;XM@##@9@
                        @@#  ,  ::.rs;isiiH@iG@#@;
                        :@sB  . ;:;r:,rs5B@@h#@X#
                         @&amp;S: .  . ,X.h##@@@@H@r@r</font><!--End left leg--></td>
<td height="1" colspan="1" width="1" valign="top"><!--Start Center Area-->
<center><table width="1" height="1" border="0" cellspacing="0" cellpadding="0">
<!--Start Title Row--><tr><td><!--Begin Main Title, Use menu.php to make these easily --><center><pre><font face="lucida console"><font size="2" color="#00FFFF"> _______________________________________________________ </font>
<font size="2" color="#00FFFF">|</font><font size=2>    </font><font size="2" color=#FFC200> + </font><font size="2" color="#FFFFFF">w e l c o m e   t o   a   n e w   s i t e</font><font size=2 color="#FFC200"> +     </font><font size=2 color="#00FFFF">|</font>
<font size="2" color=#00FFFF> ------------------------------------------------------- </font></font>
</pre></center><!--End Main Title-->
</td></tr><!-- End title Row -->
<!--Start Blurb Row--><tr><td><!--Start Blurb -->
<font color="#FFFFFF" face="lucida console" size="1">Hello There.<br>
Welcome to an experiment in web design. A truly new and revolutionary method of navigation, presentation, and creative spelling. I call it, Llama Legs.</font><!-- End Blurb-->
</td></tr><!--End Blurb Row-->
<!--Start Example News 1 -->
<tr><td>
<br><!--Start News 1 Banner -->
<pre><font face="lucida console"><font size="2" color="#ff3333"> _____________________________________________ </font>
<font size="2" color="#ff3333">|</font><font size="2"> </font><font size="2" color="#888888"> + </font><font size="2" color="#FFFFFF"><!--Start News 1 Title -->flaming wombats!<!--End News 1 Title --></font><font size="2" color="#888888"> +     </font>    <font size="2" color="#888888"> + </font><font size="2" color="#FFFFFF"><!--Start News 1 Date -->3.24.2001<!--End News 1 Date --></font><font size="2" color="#888888"> +<font size="2" color="#ff3333">|</font>
<font size="2" color="#ff3333"> --------------------------------------------- </font></font></font></pre><!--End News 1 Banner -->
</td></tr>

<tr><td><!-- Start News 1 Content -->
<font color="#FFFFFF" face="lucida console" size="1">
Good God! In recent news i was just attacked by a pair of flying wombats! They were out to kill i tell y a!  I'm pretty sure it was the meta refresh that drove 'em over the edge... Note to self: be more careful with dangerous web tags.<!-- End News 1 Content -->
</font>
</td></tr>
<!--End Example News 1 -->
<!--Start Example News 2 -->
<tr><td>
<br>
<pre><font face="lucida console"><font size="2" color="#ff3333"> _________________________________________________ </font>
<font size="2" color="#ff3333">|</font><font size="2"> </font><font size="2" color="#888888"> + </font><font size="2" color="#FFFFFF">Firey Locust Swarms!</font><font size="2" color="#888888"> +     </font>    <font size="2" color="#888888"> + </font><font size="2" color="#FFFFFF">3.22.2001</font><font size="2" color="#888888"> +<font size="2" color="#ff3333">|</font>
<font size="2" color="#ff3333"> ------------------------------------------------- </font></font></font></pre>
</td></tr>

<tr><td>
<font color="#FFFFFF" face="lucida console" size="1">
Good God! In recent news i was just attacked by a swarm of firey locusts!! I'm really not sure why they were on fire, but they looked mighty mad! I think they may have tried to use windows...oh no! We could have a crazy epidemic on our hands, if the locusts start using windows.
</font>
</td></tr>
<!--End Example News 2 -->
<!--Start Example News 3 -->
<tr><td>
<br>
<pre><font face="lucida console"><font size="2" color="#ff3333"> ___________________________________________ </font>
<font size="2" color="#ff3333">|</font><font size="2"> </font><font size="2" color="#888888"> + </font><font size="2" color="#FFFFFF">FOR GREAT HIGH</font><font size="2" color="#888888"> +     </font>    <font size="2" color="#888888"> + </font><font size="2" color="#FFFFFF">3.20.2001</font><font size="2" color="#888888"> +<font size="2" color="#ff3333">|</font>
<font size="2" color="#ff3333"> ------------------------------------------- </font></font></font></pre>
</td></tr>

<tr><td>
<font color="#FFFFFF" face="lucida console" size="1">
[Gliebster] ALL YOUR CRACK ARE BELONG TO KAT<br>
[girlsike] yesssssssssssss<br>
[Lyon] YOU HAVE NO CHANCE TO SNORT MAKE YOUR LINE<br>
[Gliebster] SET UP US THE CRACKPIPE<br>
[girlsike] SNORT OUT EVERY ZIG<br>
[Lyon] FOR GREAT HIGH 
</font>
</td></tr>
<!--End Example News 3 -->
</table></center>
</td><!-- End Center Area --> <!-- Start Right Leg-->
<td><pre><font color="#00FF00" face="lucida console" size="1">G@5.&amp;Ssr@@,  @@@r;r 2 r#293B@H5GG2XA#XX9@@.          
 ,@H,@#M@@@   @@2 r: &amp; X@25AM@2S&amp;29M@HHM@@            
  h@ @&amp;9@@    @@ :sr.G,HAi2#@hsGX3#@A&amp;A#@:            
  .@ @9M@5    @ .rsrHS,@2SH@9rX93#@BGAH@A             
   @ @h@@    ;A rsSrA:.@ih@X;Shh@@MhAH@@              
   5hH@@s   .,,;ss;,S GB3@2;i3H@@#G&amp;H@@               
    @:@@    ,:;:,. ::S@9@X:s2#@##&amp;AH#@;               
    #s@G    ,:;;,.;2r@A#X:r2@@##AAH&amp;@G                
    .h@  .  ,:;sii9;@@#2:;5@@##MAG2#@                 
    ;h@  ..:;rrsrsi#@Mr;;2@@#@@h22H@                  
    AH     .;sSSr@#@M:;:X@@#@@SiX&amp;@r                  
    @,      .:;:AH@@,rrM@@##@ii3h#M                   
    @r      .:,;#X@s;s;@@@#@SSGAA@X                   
    @s      .,.:Ai@5sr;@@@#@S2GG&amp;@S                   
    @s      ., :2r@Xr;;@#@#@s2hhG@r                   
    @2      ., :S;@hr;r@#@#@rXGG&amp;@;                   
    @H      .. rS:@&amp;;i@#@#@rXhGG@;                   
    @G      ,. rS:@A:;5@#@@@rXhhG@:                   
    @i     .,  sS:@A:;2@#@@@sX9h9@,                   
   ;Xi,    ..  si;@G:;3@#@@@iX993@,                   
   G;sr    ..  rir@9::G@#@@#52339@.                   
   # ;5    .   riS@5::A@#@@M5233H@                    
   @ ,&amp;    .   ;52@i::H@#@@H223XM@                    
   @  #   .    ;G3@i::H##@@H22Xh#@                    
   @  @   .    :H3@r;:H@#@@A322HM@                    
   @  @        :MXhr;;B@#@@&amp;XS2#M@                    
   @  @        :M5Si;;B@#@@hXS3#B@                    
   @  @      . ;#rSS;:#@M@@X3s&amp;#B@                    
   @  @      . ;A:2S;:@@#@@29rBMM@                    
  ,@  @      . ;h,2i;;@#M@@2&amp;r#M#@                    
  i@  @      . r&amp;:2S:;@##@AXAr#H#@                    
  2@  @      . rA;XS:s@##@39Ar#A#@                    
  M@  @     .. rA:X2:3@#@@XGAr#G#@                    
  B&amp;  @     ...r9:32 H@#@@X&amp;ArM&amp;#@                    
  #G  @2s   ...rA:Xr i@@@A9&amp;AsMA@@                    
  M9  @,@   ...rH,Si,;#@@2AGASB&amp;@#                
  @3  @.@   ...rH 3A::#@@iA&amp;A5AG@H                    
  B3  @:@   ...rA @#.:#@@iAGH5h3#&amp;                
  32  @ #:  .. 2G @B rH@HiHhH233#3                    
 .22  @ Bs  .  @rS@H h3@3SH9M339@2                    
 ;S5. @ G3     @:r@h M2@X2B3#hX3@i                    
 sii. @ &amp;A    ;@;5@i #2@32B9@hiX@r                    
 2ir, @ hM    9@2s@r.B9@h2H9@2r2@;                    
 2i;, @ h#   ,AH&amp;@r;GG@GXBh@2;5@:                    
 rA;, @,hM  .;BA&amp; @3rXGMG3;H@2A@@                     
  BiB  @h@ ;::,5@ M@r3&amp;BX3rXBSA@@                     
  .Hs: B&amp;@ rr, X@,S@SA#Mi3H&amp;h9#@h                     
   92&amp;  #@ ,i. s@SS@GM@Hs2hA&amp;##@.                     
    Msi @@  r:.r3@&amp;#MM@Ss2X9@H@@                      
    rGG X@s ;rrsh@&amp;G@@&amp;:si3@@A@@                      
     BA  @H ,:;2MB&amp;G@@;:rS@@HA@:                      
      @r @M .,;9H9@M@s.,;@@@99@                       
      #&amp; @#  .r&amp;AhS@5 :s@@@#s@@                       
      ,@ ,@   sh99;X ,#@@@@sG@X                       
       @: @,  s23A,  r@@@@H5#@                        
       .@ @@  ;S&amp;H  h@@@@@2M@@                        
        M r@   2@   @@22@2;@@B
         # @;  A@ ;@@@:s2X;M@s
          :H.     2@..i&amp;rA2MXM;  .</font></td><!--End Right Leg--></tr><!-- End Upper Leg/Content Area-->
<!-- Start Foot Area--><tr><td colspan="3"><pre><font color="#00FF00" face="lucida console" size="1">                        r@;@   . ;AX@@@@@@@@&amp;@;@@@@@r                                                                                             r        @  :@3S29X2@@X ,                
                          @#XG   5#X:M@##@@@@hB2@&amp;AM@@@@@.                                                                                          ,;   ., XX  @#M@;A35@H@;                
                          :@@@    #S:GH@@@@@@iAGHAAAHAH#@@@@3                                                                                   .i@@# &amp;  ,  ;;: @sh@9X9X#@#@@               
                           2X@,   @2   ;#@#X#M9HHHHHHHAAAAH@@@@@r                                                                            .2Bhr    ;; .  @:G ;.:##9@;i@@@@@3             
                        ,:       :@   A  :rA@@2BAAHHAAAHAAAAAAM@@@@#                                                         : ..     .,.  :rr.        #   s@9S  ;;ri@##,h@M#@@@.           
                       :.  ;     r:   @:.#;:##9@@@@@##MHHHAAHHA&amp;AH@@@@@:                                                    . ,:.  :r:.  .,.            &amp;  #@A2&amp; ;rr;3h@S @#M#@@@@          
                     ,:          ;    @@ 3;A@Hs;rSi5hB@@@@@#MBMA&amp;AAAHM@@                                                    ..,.  ,i                    S ,@S@Xh .rrrr5@@:rH####@@@s        
                    ,.          .    #@XS#.2#HMMA2;;;;:;s5hAM@@@HAAAA&amp;H@                                                   ..,,  .s                      ;M@,Xrr@ ,rsrr&amp;@@:;rH###@@@@       
                  .:.      ,    .    M2@@r &amp;&amp;#BM@BBA9X2is;;;;;ri#BHAAHH@@                                                  ,..  .r                      ,,A@   @5M :ssrs@@#93:rB@##@@@M     
                 ,,       ,          #S;Ai5&amp;###M######MMMHH3riirH#BHHA#@                                                  ,..  r                   @@@@@@@@   @.@s rrr,#@@#M@h:rM@##@@@r   
                :.        :         ;A&amp;S,@@53@@rrrssiiSS5223#MSiSr2#MHHH@M                                                 ,.  ;                  9@@@@@BGM@,  :5@@ ;;:.M@@s;5@@G,s#@#@@@@  
              .,         .          2GB:@@@hB@r;siSSS55522222h#9iSssMMHH@@                                                 ,  :                 @@@@@@#AM@#@   M@B ,r:;@@B;52sri#@h:i###@@@s
             ,.                     3&amp;r @3#@@@HHAh2SSS55522222XHM2SirA#HH@s                                                . :               2@@@@##@@@@#@@  .&amp;@S ,r:;@@#r;S2X2irS@@9;X###@B
            ,                      .5Hr r .i2.:XS;rsiiSS522222223M&amp;2Sr2@M@@                                                .,              @@@@@@@@@@@#HXM  #A@; ;r:s@@M#@2;S225is;S@@h;9###
          ..                     . ;i@,.A;;@9Gh&amp;#@#&amp;5ssiS555222X22GB32si##@,                                               ,            ;@@@@#@@@@@MAh2:5B.h9@  ;r:i@@MBB#@&amp;siiiis;:r@@hr&amp;#
         .                      .s HH@, A2         S@@@9iiiS5222222XH&amp;3isG@@                                              ..          H@@@@@@@@Hs;;;ri;s@BS9@  ;r:2@@MBMMB#@Mr5322Ssr:.;M@35
        .               S;       ,,  A@X r9S#23&amp;B@5   .h@@M2SiS5222229A32s3@                                              .         @@@#@@@@@@#@HhA#r i@9r3B .;r;X@@MBBBBBM#@Mr5X3X22Ss;,r#@
                      ,39             s@@r .rrs:..rXG2.   :B@@G55222229hX2SM@                                                     @@@H#@@@@MrH#@@@#G;i@9rGS .;rr&amp;@@##MBMMMMMM@@;r2X3X2Sir;:s
                     Si2                H@@r .52Ss;:,,;r&amp;2    s@@93Ah253H3SX@                                                    @@#M@@@@;       ,@ir@2i&amp; ,;ri&amp;Ah&amp;H##MMBBMMMM@@r;iS525SSssi
  -               .SSr;.                  sHH5G2AMB##HB2rrS23S.  .:.r25SSGH2@H                                             @2  @@@HhSs             2@i;s. ;SSh&amp;255S5XGHMBMMMMM#@@3i55SSiisss
</font></pre>
</td></tr><!-- End Foot Area-->
</table><!-- End Leg/Content Area -->
</center>
</td></tr>
<!-- ## END UPPER AREA ## -->
<!-- ## BEGIN BOTTOM MENU ## -->
<tr><td width="100%"><!-- Begin Menu Core: Use menu.php to make these quickly -->
<center><pre><font face="lucida console">
<font size="2" color="#00FFFF"> ___________________________________________________________________________________________________________________________________________________________________________ </font>
<font size="2" color="#00FFFF">|</font><font size="2">    </font><font size="2" color="#FFC200"> :: </font><font size="2" color="#FFFFFF"><a href="about.html">a b o u t</a></font><font size="2" color="#FFC200"> ::     </font><font size="2" color="#00FFFF">|</font><font size="2">    </font><font size="2" color="#FFC200"> :: </font><font size="2" color="#FFFFFF"><a href="index2.html">n e w s</a></font><font size="2" color="#FFC200"> ::     </font><font size="2" color="#00FFFF">|</font><font size="2">    </font><font size="2" color="#FFC200"> :: </font><font size="2" color="#FFFFFF"><a href="photo.html">p h o t o s</a></font><font size="2" color="#FFC200"> ::     </font><font size="2" color="#00FFFF">|</font><font size="2">    </font><font size="2" color="#FFC200"> :: </font><font size="2" color="#FFFFFF"><a href="comp.html">c o m p u t e r s</a></font><font size="2" color="#FFC200"> ::     </font><font size="2" color="#00FFFF">|</font><font size="2">    </font><font size="2" color="#FFC200"> :: </font><font size="2" color="#FFFFFF"><a href="index2.html">p r o g r a m m i n g</a></font><font size="2" color="#FFC200"> ::     </font><font size="2" color="#00FFFF">|</font><font size="2">    </font><font size="2" color="#FFC200"> :: </font><font size="2" color="#FFFFFF"><a  href="index2.html">a r t</a></font><font size="2" color="#FFC200"> ::     </font><font size="2" color="#00FFFF">|</font>
<font size="2" color="#00FFFF"> --------------------------------------------------------------------------------------------------------------------------------------------------------------------------- </font>
<font face="lucida console"></pre></center><!-- End Menu Core -->
</td></tr>
<!-- ## END BOTTOM MENU ## -->
</table>
<!-- ## END ALL ## -->
</center>
`

function App() {
  return (
    <BrowserRouter>
      <Header />
      { <div dangerouslySetInnerHTML={{ __html: DOMPurify.sanitize(rawHTML) }} /> }
    </BrowserRouter>
  );
}

export default App;
