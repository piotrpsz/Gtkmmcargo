# Gtkmmcargo
Own make replacement for projects (written in c ++) using gtkmm. The program compiles files (cc, cpp etc.) and builds an executable file. Program written in Go.

A program called without parameters reads the configuration from the gtkmmcargo.cfg file.<br>
If the contents of the file seem to be correct, compilation and creation of the executable file are performed.<br>
If the file does not exist, the program exits.<br><br>

The parameter allowing to work with the configuration file is the '-cfg' flag.<br><br>
<b>Example of use:</b><br>
<ul>
  <li>
    gtkmmcargo -cfg template<br>
    <i>creates an empty cfg file on the disk in the current directory, in which the user can enter relevant data,</i>
  </li>
  <li>
    gtkmmcargo -cfg cfg_file_name<br>
    <i>allows to build the project using configuration file with custom name,</i>
  </li>
  <li>
    gtkmmcargo -cfg (without parameter)<br>
    <i>displays the default flags for gtkmm.</i>
  </li>
  </ul>
  
